package crawler_test

import (
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/crawler"
)

var testFormData = &browserk.FormData{
	UserName:          "testuser",
	Password:          "testP@assw0rd1",
	FirstName:         "Test",
	MiddleInitial:     "V",
	MiddleName:        "Vuln",
	LastName:          "User",
	FullName:          "Test User",
	Address:           "99 W. 3rd Street",
	AddressLine1:      "Apt B",
	AddressLine2:      "Line 2 addr",
	AddressLineExtra:  "",
	StatePrefecture:   "CA",
	Country:           "USA",
	ZipCode:           "90210",
	City:              "Beverly Hills",
	NameOnCard:        "Test User",
	CardNumber:        "4242424242424242",
	CardCVC:           "434",
	ExpirationMonth:   "12",
	ExpirationYear:    "2022",
	Email:             "testuser@test.com",
	PhoneNumber:       "5055151",
	CountryCode:       "+1",
	AreaCode:          "555",
	Extension:         "9024",
	PassportNumber:    "20942422424",
	TravelOrigin:      "NRT",
	TravelDestination: "GCM",
	Default:           "browserker",
	SearchTerm:        "browserker",
	CommentTitle:      "browserker",
	CommentText:       "why yes indeed",
	DocumentName:      "file.txt",
	URL:               "https://example.com/browserker",
	Network:           "192.168.1.1",
	IPV4:              "192.168.1.20",
	IPV6:              "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
}

func TestFormContext(t *testing.T) {
	formHandler := crawler.NewCrawlerFormHandler(testFormData)
	mockForm := mock.MakeMockAddressForm()
	//formContext := formHandler.CreateFormContext(mockForm)
	formHandler.Fill(mockForm)
	for _, ele := range mockForm.ChildElements {
		if ele.Type == browserk.INPUT {
			t.Logf("value set (name=%s): %s\n", ele.GetAttribute("name"), ele.Value)
		}
	}
}
