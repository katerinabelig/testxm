package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/users", makeHTTPHandleFunc(s.handleUsers))
	router.HandleFunc("/company", makeHTTPHandleFunc(s.handleCompany))
	router.HandleFunc("/company/{id}", makeHTTPHandleFunc(s.handleGetCompanyByID))
	//log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleCompany(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return s.handleGetCompany(w, r)
	case http.MethodPost:
		withJWTAuth(func(w http.ResponseWriter, r *http.Request) {
			s.handleCreateCompany(w, r)
		})(w, r)
		return nil
	case http.MethodDelete:
		return s.handleDeleteCompany(w, r)
	case http.MethodPatch:
		return s.handlePatchCompany(w, r)
	default:
		return fmt.Errorf("unsupported method: %s", r.Method)
	}
}

func (s *APIServer) handleUsers(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return s.handleGetUsers(w, r)
	case http.MethodPost:
		return s.handleCreateUser(w, r)
	default:
		return fmt.Errorf("unsupported method: %s", r.Method)
	}
}

func (s *APIServer) handleGetUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := s.store.GetUsers()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, users)
}

func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateUserAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	user, err := NewUserAccount(req.Username, req.Password, req.Role)
	if err != nil {
		return err
	}
	if err := s.store.CreateUser(user); err != nil {
		return err
	}

	token, err := createJWT(user)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Token:    token,
		Username: user.Username,
	}
	return WriteJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleGetCompany(w http.ResponseWriter, r *http.Request) error {
	companies, err := s.store.GetCompanies()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, companies)
}

func (s *APIServer) handleGetCompanyByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	switch r.Method {
	case http.MethodGet:
		company, err := s.store.GetCompanyByID(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, company)

	case http.MethodDelete, http.MethodPatch:
		withJWTAuth(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				s.handleDeleteCompany(w, r)
			} else if r.Method == http.MethodPatch {
				s.handlePatchCompany(w, r)
			}
		})(w, r)
		return nil

	default:
		return fmt.Errorf("method not allowed %s", r.Method)
	}
}

func (s *APIServer) handleCreateCompany(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateCompanyRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	if req.Name == "" {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "name is required"})
	}
	if req.CompanyType == "" {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "companyType is required"})
	}
	if req.EmployeesAmount <= 0 {
		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: "employee amount should be greater than 0"})
	}
	company, err := NewCompany(req.Name, req.Description, string(req.CompanyType), req.EmployeesAmount, req.Registered)
	if err != nil {
		return err
	}
	if err := s.store.CreateCompany(company); err != nil {
		return err
	}

	idInBytes, err := json.Marshal(company.ID)
	if err != nil {
		return err
	}
	fmt.Println(req)

	err = PushCompanyEventToQueue("company_created", idInBytes)
	if err != nil {
		fmt.Println("Error pushing company_created to queue", err)
		return err
	}
	return WriteJSON(w, http.StatusOK, company)
}

func (s *APIServer) handleDeleteCompany(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	if err := s.store.DeleteCompanyByID(id); err != nil {
		return err
	}
	idInBytes, err := json.Marshal(id)
	if err != nil {
		return err
	}
	err = PushCompanyEventToQueue("company_deleted", idInBytes)
	if err != nil {
		fmt.Println("Error pushing company event to queue", err)
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func (s *APIServer) handlePatchCompany(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]

	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, fmt.Sprintf("invalid UUID format: %v", err), http.StatusBadRequest)
		return nil
	}

	company, err := s.store.GetCompanyByID(id)
	if err != nil {
		return err
	}

	updatedCompany := new(Company)
	if err := json.NewDecoder(r.Body).Decode(updatedCompany); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil
	}

	if updatedCompany.Name != "" {
		company.Name = updatedCompany.Name
	}
	if updatedCompany.Description != "" {
		company.Description = updatedCompany.Description
	}
	if updatedCompany.EmployeesAmount != 0 {
		company.EmployeesAmount = updatedCompany.EmployeesAmount
	}
	if updatedCompany.Registered != company.Registered {
		company.Registered = updatedCompany.Registered
	}
	if updatedCompany.CompanyType != "" {
		company.CompanyType = updatedCompany.CompanyType
	}

	if err := s.store.UpdateCompany(company); err != nil {
		return err
	}
	idInBytes, err := json.Marshal(id)
	if err != nil {
		return err
	}
	err = PushCompanyEventToQueue("company_updated", idInBytes)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, map[string]string{"updated": id})
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if claims["role"] != "admin" {
			permissionDenied(w)
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func createJWT(user *UserAccount) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"role":      user.Role,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func ConnectProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	conn, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func PushCompanyEventToQueue(topic string, message []byte) error {
	brokers := []string{"kafka:9092"}
	producer, err := ConnectProducer(brokers)
	if err != nil {
		return err
	}

	defer producer.Close()
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}
