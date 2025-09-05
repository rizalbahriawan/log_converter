package response

type LoginResponse struct {
	IdToken  string   `json:"idToken"`
	UserInfo UserInfo `json:"userInfo"`
	RoleId   int      `json:"roleId"`
	ListMenu []string `json:"listMenu"`
}

type UserInfo struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Kode         string `json:"kode"`
	Label        string `json:"label"`
	EmployeeID   int    `json:"employeeId"`
	RoleID       int    `json:"roleId"`
	Password     string `json:"password"`
	EmployeeName string `json:"employeeName"`
}
