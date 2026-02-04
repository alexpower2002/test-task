package team

type provider struct {
}

type Info struct {
	CreatedBy   int
	TeamMembers []int
}

func (p provider) GetAllTeams() ([]Info, error) {

}
