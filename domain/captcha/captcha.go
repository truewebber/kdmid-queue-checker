package captcha

import "github.com/truewebber/kdmid-queue-checker/domain/image"

type Solver interface {
	Solve(png image.PNG) (string, error)
}
