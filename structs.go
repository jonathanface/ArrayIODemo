package main

import (
	"encoding/json"
)

type Credentials struct {
	AppKey string `json:"appKey"`
}

type Address struct {
	Street string `json:"street" form:"street" query:"street"`
	City   string `json:"city" form:"city" query:"city"`
	State  string `json:"state" form:"state" query:"state"`
	ZIP    string `json:"zip" form:"zip" query:"zip"`
}

type User struct {
	First   string  `json:"firstName" form:"firstName" query:"firstName"`
	Last    string  `json:"lastName" form:"lastName" query:"lastName"`
	SSN     string  `json:"ssn" form:"ssn" query:"ssn"`
	DOB     string  `json:"dob" form:"dob" query:"dob"`
	AppKey  string  `json:"appKey" form:"appKey" query:"appKey"`
	Address Address `json:"address" form:"address" query:"address"`
}
type UserID struct {
	ID string `json:"id"`
}
type UserToken struct {
	Token string `json:"UserToken"`
}

type CreditRequest struct {
	ID          string `json:"id"`
	ClientKey   string `json:"clientKey"`
	ProductCode string `json:"productCode"`
}
type CreditDisplay struct {
	ReportKey    string `json:"reportKey"`
	DisplayToken string `json:"displayToken"`
}
type QA struct {
	ID           string            `json:"id"`
	AppKey       string            `json:"appKey" form:"appKey" query:"appKey"`
	ClientKey    string            `json:"clientKey" form:"clientKey" query:"clientKey"`
	AuthToken    string            `json:"authToken" form:"authToken" query:"authToken"`
	AnswerValues map[string]string `json:"answers" form:"answers" query:"answers"`
}

type APIError struct {
	Error []ErrorParams `json:"error"`
}
type ErrorParams struct {
	Msg      string `json:"msg"`
	Value    string `json:"value"`
	Param    string `json:"param"`
	Location string `json:"location"`
}
type APISuccess struct {
	ClientKey string `json:"clientKey"`
	AuthToken string `json:"authToken"`
}

type AuthAnswer struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	CorrectAnswer string `json:"correctAnswer"`
}

func (answer AuthAnswer) MarshalJSON() ([]byte, error) {
	var tmp struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
	tmp.ID = answer.ID
	tmp.Text = answer.Text
	return json.Marshal(&tmp)
}

type AuthQuestion struct {
	ID      string       `json:"id"`
	Test    string       `json:"text"`
	Answers []AuthAnswer `json:"answers"`
}

type AuthTest struct {
	AuthToken string         `json:"authToken"`
	Provider  string         `json:"provider"`
	Questions []AuthQuestion `json:"questions"`
}

func (auth AuthTest) MarshalJSON() ([]byte, error) {
	var tmp struct {
		Provider  string         `json:"provider"`
		Questions []AuthQuestion `json:"questions"`
	}
	tmp.Provider = auth.Provider
	tmp.Questions = auth.Questions
	return json.Marshal(&tmp)
}
