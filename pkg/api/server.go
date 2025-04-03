package api

import (
	"bankingsystem/pkg/services"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ContentType constants
const (
	ContentTypeJSON = "application/json"
	ContentTypeXML  = "application/xml"
)

// Server represents the HTTP server for our banking API
type Server struct {
	bankingService services.BankingServiceInterface
	router         *http.ServeMux
}

// NewServer creates a new API server
func NewServer(bankingService services.BankingServiceInterface) *Server {
	server := &Server{
		bankingService: bankingService,
		router:         http.NewServeMux(),
	}

	// Register routes
	server.registerRoutes()

	return server
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Start starts the HTTP server on the specified port
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s)
}

// registerRoutes sets up all the API endpoints
func (s *Server) registerRoutes() {
	// Customer routes
	s.router.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleListCustomers(w, r)
		case http.MethodPost:
			s.handleCreateCustomer(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.router.HandleFunc("/customers/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/customers/")

		// Handle /customers/{id}/accounts pattern
		if strings.Contains(path, "/accounts") {
			id := extractIDFromPath(r.URL.Path, "/customers/")
			if r.Method == http.MethodGet {
				s.handleGetCustomerAccounts(w, r, id)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Handle /customers/{id} pattern
		id := path
		if id == "" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.handleGetCustomer(w, r, id)
		case http.MethodPut:
			s.handleUpdateCustomer(w, r, id)
		case http.MethodDelete:
			s.handleDeleteCustomer(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Account routes
	s.router.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleCreateAccount(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	s.router.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/accounts/"):]

		// Handle subresources like /accounts/{id}/deposit
		if strings.Contains(path, "/") {
			id := extractIDFromPath(r.URL.Path, "/accounts/")

			if strings.HasSuffix(r.URL.Path, "/transactions") {
				if r.Method == http.MethodGet {
					s.handleGetAccountTransactions(w, r, id)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if strings.HasSuffix(r.URL.Path, "/deposit") {
				if r.Method == http.MethodPost {
					s.handleDeposit(w, r, id)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if strings.HasSuffix(r.URL.Path, "/withdraw") {
				if r.Method == http.MethodPost {
					s.handleWithdraw(w, r, id)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else {
				http.NotFound(w, r)
			}
			return
		}

		// Handle direct account resources /accounts/{id}
		id := path
		if id == "" {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.handleGetAccount(w, r, id)
		case http.MethodDelete:
			s.handleDeleteAccount(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Transfer route
	s.router.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleTransfer(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// extractIDFromPath extracts an ID from a path with pattern /resource/{id}
func extractIDFromPath(path, prefix string) string {
	// Remove prefix (like "/customers/") and extract ID
	id := strings.TrimPrefix(path, prefix)
	// Remove any trailing segments (like "/accounts")
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	return id
}

// determineResponseFormat decides whether to use JSON or XML based on Accept header
func determineResponseFormat(r *http.Request) string {
	accept := r.Header.Get("Accept")

	// Default to JSON if no Accept header is provided
	if accept == "" {
		return ContentTypeJSON
	}

	// Check if XML is preferred
	if strings.Contains(accept, ContentTypeXML) {
		return ContentTypeXML
	}

	// Default to JSON
	return ContentTypeJSON
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

// respondWithXML writes an XML response
func respondWithXML(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", ContentTypeXML)
	w.WriteHeader(statusCode)

	if data != nil {
		w.Write([]byte(xml.Header))
		if err := xml.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding XML response: %v", err)
		}
	}
}

// respond selects the appropriate response format based on the Accept header
func respond(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	format := determineResponseFormat(r)

	if format == ContentTypeXML {
		respondWithXML(w, statusCode, data)
	} else {
		respondWithJSON(w, statusCode, data)
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Status  int    `json:"status" xml:"status"`
	Message string `json:"message" xml:"message"`
}

// respondWithError responds with an error message
func respondWithError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	err := ErrorResponse{
		Status:  statusCode,
		Message: message,
	}
	respond(w, r, statusCode, err)
}
