//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	baseURL  = "http://localhost:8080/api"
	testUser = "testuser"
	testPass = "testpassword"
	testItem = "t-shirt"

	poorUser = "poorUser"
	poorPass = "testpassword"

	lifecycleUser = "lifecycleUser"
	lifecyclePass = "testpassword"

	user1     = "user1"
	user2     = "user2"
	password  = "testpassword"
	sendCoins = 100
)

var poorUserToken string
var jwtToken string
var lifecycleUserToken string
var user1Token, user2Token string

func TestE2E_BuyMerch(t *testing.T) {
	time.Sleep(3 * time.Second)

	t.Run("Step 1: Register and Authenticate User", func(t *testing.T) {
		authData := map[string]string{
			"username": testUser,
			"password": testPass,
		}
		body, _ := json.Marshal(authData)

		resp, err := http.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		jwtToken = result["token"]
		assert.NotEmpty(t, jwtToken)
	})

	t.Run("Step 2: Check Initial Balance", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Equal(t, float64(1000), info["coins"])
	})

	t.Run("Step 3: Buy Merch", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/buy/"+testItem, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Step 4: Check Updated Balance and Inventory", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Less(t, info["coins"].(float64), float64(1000))
		assert.NotEmpty(t, info["inventory"])
	})
}

func TestE2E_InsufficientFunds(t *testing.T) {
	time.Sleep(3 * time.Second)

	t.Run("Step 1: Register and Authenticate User", func(t *testing.T) {
		authData := map[string]string{"username": poorUser, "password": poorPass}
		body, _ := json.Marshal(authData)

		resp, err := http.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		poorUserToken = result["token"]
		assert.NotEmpty(t, poorUserToken)
	})

	t.Run("Step 2: Check Initial Balance", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+poorUserToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Equal(t, float64(1000), info["coins"])
	})

	t.Run("Step 3: Spend All Coins on Purchases", func(t *testing.T) {
		items := []string{"hoody", "powerbank", "pink-hoody"}
		for _, item := range items {
			req, _ := http.NewRequest("GET", baseURL+"/buy/"+item, nil)
			req.Header.Set("Authorization", "Bearer "+poorUserToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Step 4: Attempt to Buy Another Item (Should Fail)", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/buy/t-shirt", nil)
		req.Header.Set("Authorization", "Bearer "+poorUserToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errorResp)
		resp.Body.Close()

		assert.Contains(t, errorResp["errors"], "insufficient balance")
	})

	t.Run("Step 5: Final Check - Balance and Inventory", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+poorUserToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Equal(t, float64(0), info["coins"])
		assert.NotEmpty(t, info["inventory"])
	})
}

func TestE2E_UserLifecycle(t *testing.T) {
	time.Sleep(3 * time.Second)

	t.Run("Step 1: Register and Authenticate User", func(t *testing.T) {
		authData := map[string]string{"username": lifecycleUser, "password": lifecyclePass}
		body, _ := json.Marshal(authData)

		resp, err := http.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		lifecycleUserToken = result["token"]
		assert.NotEmpty(t, lifecycleUserToken)
	})

	t.Run("Step 2: Check Initial Balance and Inventory", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+lifecycleUserToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Equal(t, float64(1000), info["coins"])
		assert.Empty(t, info["inventory"])
	})

	t.Run("Step 3: Buy Multiple Items", func(t *testing.T) {
		items := []string{"t-shirt", "cup", "pen"}
		for _, item := range items {
			req, _ := http.NewRequest("GET", baseURL+"/buy/"+item, nil)
			req.Header.Set("Authorization", "Bearer "+lifecycleUserToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Step 4: Check Updated Balance and Inventory", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+lifecycleUserToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Less(t, info["coins"].(float64), float64(1000))
		assert.NotEmpty(t, info["inventory"])
	})

	t.Run("Step 5: Send Coins to Another User", func(t *testing.T) {
		recipient := "anotherUser"

		// Register recipient
		authData := map[string]string{"username": recipient, "password": lifecyclePass}
		body, _ := json.Marshal(authData)
		resp, err := http.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		transferData := map[string]interface{}{
			"toUser": recipient,
			"amount": 50,
		}
		body, _ = json.Marshal(transferData)

		req, _ := http.NewRequest("POST", baseURL+"/sendCoin", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+lifecycleUserToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err = client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Step 6: Final Check - Balance, Inventory, and Transactions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/info", nil)
		req.Header.Set("Authorization", "Bearer "+lifecycleUserToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var info map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&info)
		resp.Body.Close()

		assert.Less(t, info["coins"].(float64), float64(1000))
		assert.NotEmpty(t, info["inventory"])

		history, ok := info["coinHistory"].(map[string]interface{})
		assert.True(t, ok)

		sent, ok := history["sent"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, sent)

		lastTransaction, ok := sent[len(sent)-1].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(50), lastTransaction["amount"])
	})
}

func TestE2E_SendCoins(t *testing.T) {
	time.Sleep(3 * time.Second)

	t.Run("Step 1: Register and Authenticate User1", func(t *testing.T) {
		authData := map[string]string{"username": user1, "password": password}
		body, _ := json.Marshal(authData)

		resp, err := http.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		user1Token = result["token"]
		assert.NotEmpty(t, user1Token)
	})

	t.Run("Step 2: Register and Authenticate User2", func(t *testing.T) {
		authData := map[string]string{"username": user2, "password": password}
		body, _ := json.Marshal(authData)

		resp, err := http.Post(baseURL+"/auth", "application/json", bytes.NewBuffer(body))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		user2Token = result["token"]
		assert.NotEmpty(t, user2Token)
	})

	t.Run("Step 3: Check Initial Balances", func(t *testing.T) {
		checkBalance(t, user1Token, 1000)
		checkBalance(t, user2Token, 1000)
	})

	t.Run("Step 4: User1 Sends Coins to User2", func(t *testing.T) {
		transferData := map[string]interface{}{
			"toUser": user2,
			"amount": sendCoins,
		}
		body, _ := json.Marshal(transferData)

		req, _ := http.NewRequest("POST", baseURL+"/sendCoin", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+user1Token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Step 5: Check Updated Balances and Transactions", func(t *testing.T) {
		checkBalance(t, user1Token, 1000-sendCoins)
		checkBalance(t, user2Token, 1000+sendCoins)

		checkTransactionHistory(t, user1Token, "sent", sendCoins)
		checkTransactionHistory(t, user2Token, "received", sendCoins)
	})
}

func checkBalance(t *testing.T, token string, expectedBalance int) {
	req, _ := http.NewRequest("GET", baseURL+"/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	resp.Body.Close()

	assert.Equal(t, float64(expectedBalance), info["coins"])
}

func checkTransactionHistory(t *testing.T, token, transactionType string, expectedAmount int) {
	req, _ := http.NewRequest("GET", baseURL+"/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)
	resp.Body.Close()

	history, ok := info["coinHistory"].(map[string]interface{})
	assert.True(t, ok)

	transactions, ok := history[transactionType].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, transactions)

	lastTransaction, ok := transactions[len(transactions)-1].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(expectedAmount), lastTransaction["amount"])
}
