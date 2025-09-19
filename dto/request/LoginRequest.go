package request

import "github.com/go-playground/validator/v10"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthenticateRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LogConverterRequest struct {
	EmployeeID   string              `json:"employee_id" validate:"required"`
	Token        string              `json:"token" validate:"required"`
	Months       []int               `json:"months" validate:"required"`
	Year         int                 `json:"year" validate:"required"`
	ProjectName  string              `json:"project_name" validate:"required"`
	RandomizeLog RandomizeLogRequest `json:"randomize_log"`
}

type RandomizeLogRequest struct {
	IsRandom    bool `json:"is_random"`
	MinDuration int  `json:"min_duration"`
	MaxDuration int  `json:"max_duration"`
}

func (input AuthenticateRequest) Validate() error {
	validate := validator.New()

	err := validate.Struct(input)

	return err
}

func (input LogConverterRequest) Validate() error {
	validate := validator.New()

	err := validate.Struct(input)

	return err
}
