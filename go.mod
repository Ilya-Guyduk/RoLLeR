module github.com/Ilya-Guyduk/RoLLeR

replace github.com/Ilya-Guyduk/RoLLeR => ./

go 1.22.5

require (
	github.com/Ilya-Guyduk/RoLLeR/plugininterface v0.0.0-20241122120219-c148b63334cf
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
)
