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

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/leosocy/proksi/pkg/proxy"
)

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
	name   string
	config *XpathParserConfig
	logger zerolog.Logger
}

func newXpathParser(name string, config *XpathParserConfig) (*xpathParser, error) {
	return &xpathParser{
		name:   name,
		config: config,
		logger: zerolog.New(os.Stderr).With().Str("module", "spider").Str("name", name).Str("parser", "xpath").Logger(),
	}, nil
}

func (p *xpathParser) parseNode(e *colly.XMLElement) (*proxy.Proxy, error) {
	ip := e.ChildText(p.config.Selector.IP)
	port := e.ChildText(p.config.Selector.Port)
	anonymity := proxy.AnonymityUnknown
	if len(p.config.Selector.Anonymity) != 0 {
		anonymity = proxy.ParseAnonymity(e.ChildText(p.config.Selector.Anonymity))
	}
	pxy, err := proxy.NewBuilder().IP(ip).Port(port).Anonymity(anonymity).Build()
	if err == nil {
		p.logger.Debug().Str("proxy", pxy.AddrPort.String()).Msg("parsed")
	} else {
		p.logger.Warn().Err(err).Msg("")
	}
	return pxy, err
}

func (p *xpathParser) HandleResponse(resp *colly.Response, collector proxy.Collector) {
	contentType := strings.ToLower(resp.Headers.Get("Content-Type"))
	isXMLFile := strings.HasSuffix(strings.ToLower(resp.Request.URL.Path), ".xml") || strings.HasSuffix(strings.ToLower(resp.Request.URL.Path), ".xml.gz")
	if !strings.Contains(contentType, "html") && (!strings.Contains(contentType, "xml") && !isXMLFile) {
		p.logger.Warn().Str("content-type", contentType).Str("url", resp.Request.URL.String()).Msg("should not handle")
		return
	}

	if strings.Contains(contentType, "html") {
		doc, err := htmlquery.Parse(bytes.NewBuffer(resp.Body))
		if err != nil {
			p.logger.Warn().Err(err).Msg("")
			return
		}

		for _, n := range htmlquery.Find(doc, p.config.Selector.Base) {
			e := colly.NewXMLElementFromHTMLNode(resp, n)
			if pxy, err := p.parseNode(e); err == nil {
				collector.Collect(pxy)
			}
		}
	} else if strings.Contains(contentType, "xml") || isXMLFile {
		doc, err := xmlquery.Parse(bytes.NewBuffer(resp.Body))
		if err != nil {
			p.logger.Warn().Err(err).Msg("")
			return
		}

		xmlquery.FindEach(doc, p.config.Selector.Base, func(i int, n *xmlquery.Node) {
			e := colly.NewXMLElementFromXMLNode(resp, n)
			if pxy, err := p.parseNode(e); err == nil {
				collector.Collect(pxy)
			}
		})
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
			return nil, errors.New("spider: regex pattern must have ip and port subexpression name, for example '(?P<ip>\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}):(?P<port>\\d{1,5})'")
		}
	}
	return parser, nil
}

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
