package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateCompany(*Company) error
	DeleteCompanyByID(string) error
	UpdateCompany(*Company) error
	GetCompanies() ([]*Company, error)
	GetCompanyByID(string) (*Company, error)

	CreateUser(*UserAccount) error
	GetUsers() ([]*UserAccount, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "host=host.docker.internal user=postgres dbname=postgres password=gocompany sslmode=disable port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to the database!")
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	if err := s.createCompanyTable(); err != nil {
		return err
	}

	if err := s.createUserTable(); err != nil {
		return err
	}

	return nil
}
func (s *PostgresStore) createUserTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
        username VARCHAR(20) NOT NULL UNIQUE,
        role VARCHAR(10) CHECK (role IN ('admin', 'user')),
        password VARCHAR(100) NOT NULL
    )`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) createCompanyTable() error {
	query := `CREATE TABLE IF NOT EXISTS company (
        id UUID PRIMARY KEY NOT NULL,
        name VARCHAR(15) NOT NULL UNIQUE,
        description VARCHAR(3000),
        employees_amount INT NOT NULL,
        registered BOOLEAN NOT NULL,
        company_type VARCHAR(50) CHECK (company_type IN ('Corporations', 'NonProfit', 'Cooperative', 'Sole Proprietorship')) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateCompany(c *Company) error {
	_, err := s.db.Exec(
		"INSERT INTO company (id, name, description, company_type, employees_amount, registered, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		c.ID, c.Name, c.Description, c.CompanyType, c.EmployeesAmount, c.Registered, c.CreatedAt)
	return err
}

func (s *PostgresStore) DeleteCompanyByID(id string) error {
	_, err := s.db.Exec("DELETE FROM company WHERE id::text = $1", id)
	return err
}

func (s *PostgresStore) UpdateCompany(c *Company) error {
	_, err := s.db.Exec(
		"UPDATE company SET name = $1, description = $2, company_type = $3, employees_amount = $4, registered = $5 WHERE id = $6",
		c.Name, c.Description, c.CompanyType, c.EmployeesAmount, c.Registered, c.ID)
	return err
}

func (s *PostgresStore) GetCompanies() ([]*Company, error) {
	rows, err := s.db.Query("SELECT id, name, description, employees_amount, registered, company_type, created_at FROM company")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companys []*Company
	for rows.Next() {
		company, err := scanIntoCompany(rows)
		if err != nil {
			return nil, err
		}
		companys = append(companys, company)
	}

	return companys, nil
}

func (s *PostgresStore) GetCompanyByID(id string) (*Company, error) {
	rows, err := s.db.Query("SELECT id, name, description, employees_amount, registered, company_type, created_at FROM company WHERE id::text = $1", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		return scanIntoCompany(rows)
	}
	return nil, fmt.Errorf("company not found")
}

func scanIntoCompany(rows *sql.Rows) (*Company, error) {
	company := new(Company)
	err := rows.Scan(
		&company.ID,
		&company.Name,
		&company.Description,
		&company.EmployeesAmount,
		&company.Registered,
		&company.CompanyType,
		&company.CreatedAt)

	return company, err
}

func (s *PostgresStore) CreateUser(c *UserAccount) error {
	_, err := s.db.Exec(
		"INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		c.Username, c.Password, c.Role)
	return err
}

func (s *PostgresStore) GetUsers() ([]*UserAccount, error) {
	rows, err := s.db.Query("SELECT username, password, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*UserAccount
	for rows.Next() {
		var username, password, role string
		if err := rows.Scan(&username, &password, &role); err != nil {
			return nil, err
		}
		users = append(users, &UserAccount{
			Username: username,
			Password: password,
			Role:     role,
		})
	}

	return users, nil
}

func (s *PostgresStore) GetUserByID(id string) (*Company, error) {
	rows, err := s.db.Query("SELECT username, name, description, employees_amount, registered, company_type, created_at FROM company WHERE id::text = $1", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		return scanIntoCompany(rows)
	}
	return nil, fmt.Errorf("company not found")
}
