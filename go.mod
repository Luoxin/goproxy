module github.com/Luoxin/goproxy

go 1.16

require (
	github.com/Luoxin/Eutamias v0.0.0-20210504035458-4ba85ac2834f
	github.com/antonfisher/nested-logrus-formatter v1.3.1
	github.com/elliotchance/pie v1.38.2
	github.com/go-resty/resty/v2 v2.6.0
	github.com/gofiber/fiber/v2 v2.8.0
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/sirupsen/logrus v1.8.1
)

replace gopkg.in/go-playground/validator.v10 => github.com/go-playground/validator/v10 v10.4.1
