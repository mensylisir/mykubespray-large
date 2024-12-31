package keycloak

import (
	"encoding/json"
	"fmt"
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"os"
	"sync"
	"time"
)

type UserBatchConfig struct {
	KeycloakConfig KeycloakConfig
	MaxConcurrent  int
	MaxRetries     int
	RetryDelay     time.Duration
	ClientTimeout  time.Duration
	TickerInterval time.Duration
	LogFilePath    string
}

type UserBatch struct {
	Config         UserBatchConfig
	keycloakClient keycloakClient
	logFile        *os.File
	logFileMutex   sync.Mutex
	logQueue       chan string
	stopLogging    int32
}

func (ub *UserBatch) performAction(action string, userID string, user KeycloakUser) {
	var respErr error
	for attempt := 0; attempt < ub.Config.MaxRetries; attempt++ {
		token, err := ub.keycloakClient.GetToken()
		if err != nil {
			logger.GetLogger().Errorf("Error getting token: %v", err)
			return
		}
		objToken := make(map[string]interface{})
		err = json.Unmarshal([]byte(token), &objToken)
		if err != nil {
			logger.GetLogger().Errorf("Error getting token: %v", err)
			return
		}
		acessToken := objToken["access_token"].(string)
		switch action {
		case "create":
			respErr = ub.keycloakClient.CreateUser(acessToken, user)
		case "delete":
			respErr = ub.keycloakClient.DeleteUser(acessToken, userID)
		case "update":
			respErr = ub.keycloakClient.UpdateUser(acessToken, userID, user)
		}
		if respErr == nil {
			logger.GetLogger().Errorf("Operation succeeded for user: %s", user.Username)
			return
		}
		logger.GetLogger().Errorf("Attempt %d: Error performing operation for user %s: %v", attempt+1, user.Username, respErr)
		time.Sleep(ub.Config.RetryDelay)
	}
	logger.GetLogger().Errorf("Failed to perform operation for user %s after %d attempts", user.Username, ub.Config.MaxRetries)
}

func (ub *UserBatch) CreateUser(user KeycloakUser, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{}
	defer func() { <-sem }()

	ub.performAction("create", "", user)
}

func (ub *UserBatch) DeleteUser(userID string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{}
	defer func() { <-sem }()

	ub.performAction("delete", userID, KeycloakUser{})
}

func (ub *UserBatch) UpdateUser(userID string, user KeycloakUser, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{}
	defer func() { <-sem }()

	ub.performAction("update", userID, user)
}

func (ub *UserBatch) ProcessUsers(users []KeycloakUser, action string) {
	var wg sync.WaitGroup
	results := make(chan string, len(users))
	sem := make(chan struct{}, ub.Config.MaxConcurrent)

	for _, user := range users {
		wg.Add(1)
		switch action {
		case "create":
			go ub.CreateUser(user, &wg, sem)
		case "delete":
			go ub.DeleteUser(user.Username, &wg, sem) // Assuming username as ID here
		case "update":
			go ub.UpdateUser(user.Username, user, &wg, sem) // Assuming username as ID here
		}
	}

	wg.Wait()
	close(results)
	for result := range results {
		fmt.Println(result)
	}
}

func main() {
	kconfig := KeycloakConfig{
		ClientCredentialsConfig: ClientCredentialsConfig{
			BaseConfig: BaseConfig{
				BaseUrl:      "https://keycloak.kmpp.io",
				Reamls:       "cars",
				ClientID:     "kubernetes",
				ClientSecret: "PeWtYhmI3xFTwRDObsi7wkAFyvwvA07t",
			},
		},
	}
	//config := UserBatchConfig{
	//	KeycloakConfig: kconfig,
	//	MaxConcurrent:  50,
	//	MaxRetries:     3,
	//	RetryDelay:     2 * time.Second,
	//	ClientTimeout:  10 * time.Second,
	//	TickerInterval: 5 * time.Second,
	//	LogFilePath:    "failed_users.log",
	//}

	client := NewKeycloakClient("client_credentials", kconfig, 10*time.Second)

	//userBatch := UserBatch{
	//	Config:         config,
	//	keycloakClient: *client,
	//}
	//
	//users := []KeycloakUser{
	//	NewUser("user1", "user1@example.com", "User", "One"),
	//	NewUser("user2", "user2@example.com", "User", "Two"),
	//	// Add more users as needed
	//}
	//userBatch.ProcessUsers(users, "create")
	token, _ := client.GetToken()
	objToken := make(map[string]interface{})
	err := json.Unmarshal([]byte(token), &objToken)
	if err != nil {
		logger.GetLogger().Errorf("Error getting token: %v", err)
		return
	}
	acessToken := objToken["access_token"].(string)
	group := GroupRepresentation{
		Name: "my-new-group3",
		Path: "/my-new-group3",
	}
	client.CreateGroup(acessToken, &group)
}
