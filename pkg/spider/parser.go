// Copyright 2023 The proksi Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package spider

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/leosocy/proksi/pkg/proxy"
)

// ProxyParser is an interface for parsing proxies from HTTP responses.
// Different implementations can be used to parse proxies from different formats.
// For example, xpathParser can be used to parse proxies from HTML or XML documents using XPath expressions,
// and regexParser can be used to parse proxies from HTML or XML documents using regular expressions.
type ProxyParser interface {
	HandleResponse(resp *colly.Response, collector proxy.Collector)
}

type XpathParserConfig struct {
	Selector struct {
		Base      string
		IP        string
		Port      string
		IPPort    string
		Anonymity string
	}
}

type RegexParserConfig struct {
	Expr struct {
		IPPort string
	}
}

type ParserConfig struct {
	Type  string
	Xpath *XpathParserConfig
	Regex *RegexParserConfig
}

// xpathParser is a ProxyParser implementation that extracts proxies from HTML or XML documents using XPath expressions.
type xpathParser struct {
	name     string
	config   *XpathParserConfig
	selector struct {
		base      *xpath.Expr
		ip        *xpath.Expr
		port      *xpath.Expr
		anonymity *xpath.Expr
	}
	logger zerolog.Logger
}

func newXpathParser(name string, config *XpathParserConfig) (*xpathParser, error) {
	var err error
	parser := &xpathParser{
		name:   name,
		config: config,
		logger: zerolog.New(os.Stderr).With().Str("module", "spider").Str("name", name).Str("parser", "xpath").Logger(),
	}
	if parser.selector.base, err = xpath.Compile(config.Selector.Base); err != nil {
		return nil, err
	}
	if config.Selector.IP != "" {
		if parser.selector.ip, err = xpath.Compile(config.Selector.IP); err != nil {
			return nil, err
		}
	}
	if config.Selector.Port != "" {
		if parser.selector.port, err = xpath.Compile(config.Selector.Port); err != nil {
			return nil, err
		}
	}
	if config.Selector.Anonymity != "" {
		if parser.selector.anonymity, err = xpath.Compile(config.Selector.Anonymity); err != nil {
			return nil, err
		}
	}
	return parser, nil
}

func (p *xpathParser) querySelectorText(top interface{}, selector *xpath.Expr) string {
	switch n := top.(type) {
	case *html.Node:
		child := htmlquery.QuerySelector(n, selector)
		if child == nil {
			return ""
		}
		return strings.TrimSpace(htmlquery.InnerText(child))
	case *xmlquery.Node:
		child := xmlquery.QuerySelector(n, selector)
		if child == nil {
			return ""
		}
		return strings.TrimSpace(child.InnerText())
	default:
		return ""
	}
}

func (p *xpathParser) parseNode(node interface{}, collector proxy.Collector) {
	builder := proxy.NewBuilder().IP(p.querySelectorText(node, p.selector.ip)).Port(p.querySelectorText(node, p.selector.port))
	if p.selector.anonymity != nil {
		builder.AnonymityString(p.querySelectorText(node, p.selector.anonymity))
	}
	pxy, err := builder.Build()
	if err == nil {
		p.logger.Debug().Str("proxy", pxy.AddrPort.String()).Msg("parsed")
		collector.Collect(pxy)
	} else {
		p.logger.Warn().Err(err).Msg("")
	}
}

func (p *xpathParser) handleHTML(resp *colly.Response, collector proxy.Collector) {
	doc, err := htmlquery.Parse(bytes.NewBuffer(resp.Body))
	if err != nil {
		p.logger.Warn().Err(err).Msg("")
		return
	}

	for _, n := range htmlquery.QuerySelectorAll(doc, p.selector.base) {
		p.parseNode(n, collector)
	}
}

func (p *xpathParser) handleXML(resp *colly.Response, collector proxy.Collector) {
	doc, err := xmlquery.Parse(bytes.NewBuffer(resp.Body))
	if err != nil {
		p.logger.Warn().Err(err).Msg("")
		return
	}

	for _, n := range xmlquery.QuerySelectorAll(doc, p.selector.base) {
		p.parseNode(n, collector)
	}
}

func (p *xpathParser) HandleResponse(resp *colly.Response, collector proxy.Collector) {
	contentType := strings.ToLower(resp.Headers.Get("Content-Type"))
	isXMLFile := strings.HasSuffix(strings.ToLower(resp.Request.URL.Path), ".xml") || strings.HasSuffix(strings.ToLower(resp.Request.URL.Path), ".xml.gz")
	if !strings.Contains(contentType, "html") && (!strings.Contains(contentType, "xml") && !isXMLFile) {
		p.logger.Warn().Str("content-type", contentType).Str("url", resp.Request.URL.String()).Msg("should not handle")
		return
	}

	if strings.Contains(contentType, "html") {
		p.handleHTML(resp, collector)
	} else if strings.Contains(contentType, "xml") || isXMLFile {
		p.handleXML(resp, collector)
	}
	return
}

// regexParser is a ProxyParser implementation that extracts proxies from HTML or XML documents using regular expressions.
type regexParser struct {
	name   string
	config *RegexParserConfig
	ipport struct {
		regex   *regexp.Regexp // regular expression to match IP and port
		ipIdx   int            // index of IP subexpression in the regular expression
		portIdx int            // index of port subexpression in the regular expression
	}
	logger zerolog.Logger
}

// HandleResponse extracts proxies from the response body using the regular expression defined in the configuration.
func (p *regexParser) HandleResponse(resp *colly.Response, collector proxy.Collector) {
	matches := p.ipport.regex.FindAllSubmatch(resp.Body, -1)
	for _, match := range matches {
		ip := string(match[p.ipport.ipIdx])
		port := string(match[p.ipport.portIdx])

		pxy, err := proxy.NewBuilder().IP(ip).Port(port).Build()
		if err == nil {
			p.logger.Debug().Str("proxy", pxy.AddrPort.String()).Msg("parsed")
			collector.Collect(pxy)
		} else {
			p.logger.Warn().Err(err).Msg("")
		}
	}
}
func newRegexParser(name string, config *RegexParserConfig) (*regexParser, error) {
	var err error
	parser := &regexParser{
		name:   name,
		config: config,
		logger: zerolog.New(os.Stderr).With().Str("module", "spider").Str("name", name).Str("parser", "regex").Logger(),
	}
	if config.Expr.IPPort != "" {
		if parser.ipport.regex, err = regexp.Compile(config.Expr.IPPort); err != nil {
			return nil, err
		}
		parser.ipport.ipIdx = parser.ipport.regex.SubexpIndex("ip")
		parser.ipport.portIdx = parser.ipport.regex.SubexpIndex("port")
		if parser.ipport.ipIdx == -1 || parser.ipport.portIdx == -1 {
			return nil, errors.New("spider: regex pattern must have ip and port subexpression name, for example '(?P<ip>\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}):(?P<port>\\d{1,5})'") //nolint:lll
		}
	}
	return parser, nil
}

// NewProxyParser creates a new ProxyParser based on the given configuration.
// It returns an error if the configuration is invalid or the parser type is unsupported.
func NewProxyParser(name string, config *ParserConfig) (ProxyParser, error) {
	switch config.Type {
	case "xpath":
		return newXpathParser(name, config.Xpath)
	case "regex":
		return newRegexParser(name, config.Regex)
	default:
		return nil, errors.New(fmt.Sprintf("spider: unsupported parser type: %s", config.Type))
	}
}
