package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username int64  `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type CreateUserAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type CreateCompanyRequest struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	CompanyType     CompanyType `json:"companyType"`
	EmployeesAmount int64       `json:"employeesAmount"`
	Registered      bool        `json:"registered"`
}

type CompanyType string

const (
	Corporations       CompanyType = "Corporations"
	NonProfit          CompanyType = "NonProfit"
	Cooperative        CompanyType = "Cooperative"
	SoleProprietorship CompanyType = "Sole Proprietorship"
)

type UserAccount struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type Company struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	EmployeesAmount int64     `json:"employeesAmount"`
	Registered      bool      `json:"registered"`
	CompanyType     string    `json:"companyType"`
	CreatedAt       time.Time `json:"createdAt"`
}

func NewCompany(name, description, companyType string, employeesAmount int64, registered bool) (*Company, error) {
	switch string(companyType) {
	case string(Corporations), string(NonProfit), string(Cooperative), string(SoleProprietorship):
	default:
		return nil, errors.New("invalid company type")
	}

	id := uuid.New().String()
	return &Company{
		ID:              id,
		Name:            name,
		Description:     description,
		EmployeesAmount: employeesAmount,
		Registered:      registered,
		CompanyType:     companyType,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

func NewUserAccount(username, password, role string) (*UserAccount, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &UserAccount{
		Username: username,
		Password: string(encpw),
		Role:     role,
	}, nil
}
