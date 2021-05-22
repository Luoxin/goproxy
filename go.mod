module github.com/Luoxin/goproxy

go 1.16

require (
	github.com/Luoxin/Eutamias v0.0.0-20210507111911-049c0ad8bdb3
	github.com/antonfisher/nested-logrus-formatter v1.3.1
	github.com/elliotchance/pie v1.38.2
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/go-resty/resty/v2 v2.6.0
	github.com/gofiber/fiber/v2 v2.9.0
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf // indirect
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed // indirect
	golang.org/x/sys v0.0.0-20210507161434-a76c4d0a0096 // indirect
)

replace gopkg.in/go-playground/validator.v10 => github.com/go-playground/validator/v10 v10.4.1
