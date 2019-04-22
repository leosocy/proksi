module github.com/Leosocy/gipp

require (
	github.com/EDDYCJY/fake-useragent v0.2.0
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/Sirupsen/logrus v1.0.6
	github.com/elazarl/goproxy v0.0.0-20190410145444-c548f45dcf1d // indirect
	github.com/magiconair/properties v1.8.0
	github.com/mdempsky/gocode v0.0.0-20190203001940-7fb65232883f // indirect
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/parnurzeal/gorequest v0.2.15
	github.com/pkg/errors v0.8.1 // indirect
	github.com/skratchdot/open-golang v0.0.0-20190402232053-79abb63cd66e // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.2.0
	github.com/stretchr/testify v1.3.0
)

replace (
	golang.org/x/net v0.0.0-20180218175443-cbe0f9307d01 => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/net v0.0.0-20181114220301-adae6a3d119a => github.com/golang/net v0.0.0-20190404232315-eb5bcb51f2a3
)

replace golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2 => github.com/golang/crypto v0.0.0-20190411191339-88737f569e3a
