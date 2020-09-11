package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

const (
	ArrayUserURL         = "https://sandbox.array.io/api/user/v2"
	ArrayAuthURL         = "https://sandbox.array.io/api/authenticate/v2"
	ArrayReportURL       = "https://sandbox.array.io/api/report/v2"
	APIPrefix            = "/api"
	TuiProductReportCode = "tui1bReportScore"
)

var authToken string
var userTokens = make(map[string]string)
var clientKeys = make(map[string]string)
var credentials = new(Credentials)

/**
 * getConfiguration
 *
 * Loads in the external configuration json file,
 * or panics if there's a problem.
 */
func getConfiguration() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &credentials)
	jsonFile.Close()
}

/**
 * getUserQuestions
 *
 * Retrieves a series of peronally-identifiable questions
 * to be displayed to the authentication requestor.
 */
func getUserQuestions(c echo.Context) error {
	u := new(UserID)
	if err := c.Bind(u); err != nil {
		return err
	}
	url := ArrayAuthURL + "?appKey=" + credentials.AppKey + "&clientKey=" + clientKeys[u.ID] + "&provider1=tui&provider2=exp&provider3=efx"
	log.Println("Requesting questions from: " + url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	//log.Println("questions body", string(body))
	if resp.StatusCode == http.StatusOK {
		authTest := AuthTest{}
		json.Unmarshal(body, &authTest)
		authToken = authTest.AuthToken
		//remarshaling to strip out correctanswer fields & authtoken
		toJSON, err := json.Marshal(authTest)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(resp.StatusCode, string(toJSON))
	}
	return c.String(resp.StatusCode, string(body))
}

/**
 * createUser
 *
 * Creates a new user in the system who can then
 * have credit reports generated.
 */
func createUser(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return err
	}
	log.Println("cred", credentials.AppKey)
	u.AppKey = credentials.AppKey
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(u)
	req, _ := http.NewRequest("POST", ArrayUserURL, buf)
	req.Header.Set("Content-Type", "application/json")
	log.Println("u", u)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("err", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		apiErr := APIError{}
		json.Unmarshal(body, &apiErr)
		toJSON, err := json.Marshal(apiErr)
		if err != nil {
			log.Println("err", err.Error())
			return c.String(http.StatusBadRequest, err.Error())
		}
		c.String(resp.StatusCode, string(toJSON))
	} else {
		apiSuccess := APISuccess{}
		json.Unmarshal(body, &apiSuccess)
		log.Println("Received client key: " + apiSuccess.ClientKey)
		clientKeys[u.SSN] = apiSuccess.ClientKey
		//clientKey = apiSuccess.ClientKey
		c.String(resp.StatusCode, "")
	}
	return nil
}

/**
 * checkAnswers
 *
 * Submits a map of question and answer keypairs and
 * returns success or failure.
 */
func checkAnswers(c echo.Context) error {
	qa := new(QA)
	if err := c.Bind(qa); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	log.Println("ssn", qa.ID)
	qa.AppKey = credentials.AppKey
	qa.AuthToken = authToken
	qa.ClientKey = clientKeys[qa.ID]
	log.Println("qa", qa)

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(qa)
	req, _ := http.NewRequest("POST", ArrayAuthURL, buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Charset", "utf-8")
	//writePayload(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("qa body", string(body))
	if resp.StatusCode == http.StatusOK {
		uToken := new(UserToken)
		json.Unmarshal(body, &uToken)
		userTokens[qa.ID] = uToken.Token
		return c.String(resp.StatusCode, "")
	}
	return c.String(resp.StatusCode, string(body))
}

/* 
 * requestReport
 * 
 * Initiates a credit report for the passed user
 * The remote service will return code 206 until
 * the report is ready for display.
 */
func requestReport(c echo.Context) error {
	//"clientKey": "test-customer-123",
	//"productCode": "tui1bReportScore"
	cr := new(CreditRequest)
	if err := c.Bind(cr); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	cr.ClientKey = clientKeys[cr.ID]
	cr.ProductCode = TuiProductReportCode
	log.Println("ck", cr.ClientKey)
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(cr)
	req, _ := http.NewRequest("POST", ArrayReportURL, buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-credmo-user-token", userTokens[cr.ID])
	log.Println("id", cr.ID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	log.Println("request body", string(body))
	return c.String(resp.StatusCode, string(body))
}

/**
 * deauthUser
 * 
 * Removes the created user from the (local) dictionary
 * This will require a user to be reauthenticated if credit
 * info is later requested.
 */
func deauthUser(c echo.Context) error {
	u := new(UserID)
	if err := c.Bind(u); err != nil {
		return err
	}
	delete(clientKeys, u.ID)
	delete(userTokens, u.ID)
	return c.String(http.StatusOK, "")
}

/**
 * viewReport
 *
 * Returns the json dump of the requested
 * credit report. Contingent upon a prior
 * successful request to requestReport
 */
func viewReport(c echo.Context) error {
	cd := new(CreditDisplay)
	if err := c.Bind(cd); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	log.Println(cd)
	url := ArrayReportURL + "?reportKey=" + cd.ReportKey + "&displayToken=" + cd.DisplayToken
	log.Println("url", url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	log.Println("stat?", resp.StatusCode)
	log.Println("report body", string(body))
	return c.String(resp.StatusCode, string(body))
}

/**
 * writePayload
 *
 * Debug function to dump out json requests in an http.Client
 * Param req: the request you want to exam
 */
func writePayload(req *http.Request) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%q", dump)
}

func main() {
	getConfiguration()
	e := echo.New()
	e.Static("", "static")

	e.POST(APIPrefix+"/user", createUser)
	e.PUT(APIPrefix+"/user/revoke", deauthUser)
	e.GET(APIPrefix+"/user/questions", getUserQuestions)
	e.PUT(APIPrefix+"/qa", checkAnswers)
	e.PUT(APIPrefix+"/report/request", requestReport)
	e.GET(APIPrefix+"/report/view", viewReport)
	e.Logger.Fatal(e.Start(":1323"))
}
