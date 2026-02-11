package member

type Role string

const OwnerRole Role = "owner"
const AdminRole Role = "admin"
const NormalRole Role = "normal"

type Model struct {
	UserId int
	TeamId int
	Role   Role
}
