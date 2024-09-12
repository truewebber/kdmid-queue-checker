package page

import (
	"fmt"
	"io"

	"github.com/truewebber/kdmid-queue-checker/domain/image"
)

var ErrCaptchaNotSolved = fmt.Errorf("captcha not solved")

type Stat struct {
	HTML                 []byte
	Network              []byte
	Screenshot           image.PNG
	Captcha              Captcha
	SomethingInteresting bool
}

type Captcha struct {
	Presented bool
	Image     image.PNG
}

type Navigator interface {
	io.Closer

	OpenPageToAuthorize() (Stat, error)
	SubmitAuthorization(code string) (Stat, error)
	OpenSlotBookingPage() (Stat, error)
}

type Dispatcher interface {
	NewNavigator(id, cd string) (Navigator, error)
}
