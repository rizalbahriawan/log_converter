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

type InsertLogRequest struct {
	IdEmployee         string  `json:"idEmployee" validate:"required"`
	Token              string  `json:"token,omitempty" validate:"required"`
	ActivityDetail     string  `json:"activityDetail" validate:"required"`
	Date               string  `json:"date" validate:"required"`
	Duration           int     `json:"duration" validate:"required"`
	Months             int     `json:"months" validate:"required"`
	Overtime           int     `json:"overtime"`
	ProjectAssignId    int     `json:"projectAssignId" validate:"required"`
	ProjectId          int     `json:"projectId" validate:"required"`
	PurchaseDateString string  `json:"purchaseDateString" validate:"required"`
	SpvAssigned        *string `json:"spvAssigned"`
	SubProAssignmentId *int    `json:"subProAssignmentId"`
	SubProId           *int    `json:"subProId"`
	ThirdParty         *string `json:"thirdParty"`
	WorkingMode        string  `json:"workingMode" validate:"required"`
	Years              int     `json:"years" validate:"required"`
}
type RandomizeLogRequest struct {
	IsRandom    bool `json:"is_random"`
	MinDuration int  `json:"min_duration"`
	MaxDuration int  `json:"max_duration"`
}

type ExportParam struct {
	ProjectFilter       string
	IsRandomizeDuration bool
	MinDuration         int
	MaxDuration         int
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

func (input InsertLogRequest) Validate() error {
	validate := validator.New()

	err := validate.Struct(input)

	return err
}
