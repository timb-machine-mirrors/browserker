package crawler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	regen "github.com/zach-klippenstein/goregen"
	"gitlab.com/browserker/browserk"
)

// CrawlerFormHandler handles filling forms
type CrawlerFormHandler struct {
	formData *browserk.FormData
}

// NewCrawlerFormHandler will fill forms based on the provided formData and determining
// context for each form input
func NewCrawlerFormHandler(formData *browserk.FormData) *CrawlerFormHandler {
	return &CrawlerFormHandler{formData: formData}
}

// Init the form filler
// TODO: validate form data isn't empty etc
func (c *CrawlerFormHandler) Init() error {
	return nil
}

// InputDetails for an input tag
// TODO: Handle file
type InputDetails struct {
	Name        string
	ID          string
	Type        string
	PlaceHolder string
	AriaLabel   string
	LabelText   string
	Min         string
	Max         string
	Multiple    bool
	Required    bool
	Step        string
	Src         string
	Alt         string
	Pattern     string
	Title       string
	Unchecked   bool
}

// FormContext for auto filling easier
type FormContext struct {
	Action string
	Submit []byte                   // hash of the element we will use to submit
	Inputs map[string]*InputDetails // element hash -> ToLower'd type, name, id, place holder
}

// NewFormContext to build out a form's context
func NewFormContext(action string) *FormContext {
	return &FormContext{
		Action: action,
		Inputs: make(map[string]*InputDetails, 0),
	}
}

// AddLabel information for a particular input field, create
// a new one if it doesn't exist
func (f *FormContext) AddLabel(forHash, innerText string) {
	if _, exist := f.Inputs[forHash]; !exist {
		f.Inputs[forHash] = &InputDetails{}
	}
	f.Inputs[forHash].LabelText = strings.ToLower(innerText)
}

// AddInput to this form context
func (f *FormContext) AddInput(hash string, details *InputDetails) {
	if _, exist := f.Inputs[hash]; !exist {
		f.Inputs[hash] = details
		return
	}

	// if it does exist grab the existing label text and add it to the
	// incoming details then just overwrite
	label := f.Inputs[hash].LabelText
	if label != "" {
		details.LabelText = label
		f.Inputs[hash] = details
	}
}

// Fill the form with context relevant data
func (c *CrawlerFormHandler) Fill(form *browserk.HTMLFormElement) {
	formContext := c.CreateFormContext(form)
	for eleHash, input := range formContext.Inputs {
		ele := form.GetChildByHash([]byte(eleHash))
		ele.Value = c.GetSuggestedInput(input)
		log.Info().Msgf("suggested %s for ele %s", ele.Value, ele.GetAttribute("name"))
	}
	form.SubmitButtonID = formContext.Submit
	return
}

// GetSuggestedInput given input details, try to return a valid value
func (c *CrawlerFormHandler) GetSuggestedInput(input *InputDetails) string {
	label := input.AriaLabel + input.LabelText + input.PlaceHolder
	isStart := DateStartRe.MatchString(label)
	isEnd := DateEndRe.MatchString(label)
	now := time.Now()
	// check element type first as that's the heighest weight and will allow us to
	// exit out early
	switch input.Type {
	case "reset", "button", "submit":
		return "" // this could be a reset/clear button so do nothing
	case "checkbox":
		return ""
	case "color":
		return "#e66465"
	case "date":
		if isStart && input.Min != "" {
			return input.Min
		}

		if isEnd && input.Max != "" {
			return input.Max
		}

		if isEnd {
			now.Add(time.Hour * 72)
		}
		return fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())
	case "datetime-local", "datetime":
		if isStart && input.Min != "" {
			return input.Min
		}

		if isEnd && input.Max != "" {
			return input.Max
		}

		if isEnd {
			now.Add(time.Hour * 72)
		}
		// choose 9:30 because it's common for reservation systems to do increments of 15 minutes
		// and it's within 'business hours'
		return fmt.Sprintf("%d-%02d-%02dT10:30", now.Year(), now.Month(), now.Day())
	case "email":
		return c.formData.Email
	case "file":
	case "hidden":
		return ""
	case "image":
		// TODO: handle image based forms
		// ref: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/image
	case "month":
		if isStart && input.Min != "" {
			return input.Min
		}

		if isEnd && input.Max != "" {
			return input.Max
		}

		if isEnd {
			now.Add(time.Hour * 754) // next month-ish
		}

		return fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	case "number", "range":
		var err error
		var max, min, step, val int

		if min, err = strconv.Atoi(input.Min); err != nil {
			min = 1
		}

		if max, err = strconv.Atoi(input.Max); err != nil {
			max = 10
		}

		if step, err = strconv.Atoi(input.Step); err != nil {
			step = 1 // default step is always 1
		}
		val = min + step
		if val > max {
			val = max - step
		}
		return strconv.Itoa(val)
	case "password":
		return c.formData.Password
	case "radio":
		return ""
	case "tel":
		if input.Pattern != "" {
			result, err := regen.Generate(input.Pattern)
			if err != nil {
				return result
			}
		}
		return c.formData.PhoneNumber
	case "search", "text":
		return c.suggestTextInput(input)
	case "time":
	case "url":
		return c.formData.URL
	case "week":
		if isStart && input.Min != "" {
			return input.Min
		}

		if isEnd && input.Max != "" {
			return input.Max
		}

		if isEnd {
			now.Add(time.Hour * 160) // next week-ish
		}
		year, week := now.ISOWeek()
		return fmt.Sprintf("%d:W%02d", year, week)
	}

	return c.suggestTextInput(input)
}

// there be dragons here
func (c *CrawlerFormHandler) suggestTextInput(input *InputDetails) string {
	label := input.AriaLabel + input.LabelText + input.PlaceHolder

	id := input.Name
	if id == "" {
		id = input.ID
	}

	// search form
	if SearchTermRe.MatchString(id) || SearchTermRe.MatchString(label) {
		return c.formData.SearchTerm
	}

	// check country
	if RegionIgnoredRe.MatchString(id) || CountryRe.MatchString(id) ||
		CountryLocationRe.MatchString(id) || CountryRe.MatchString(label) {
		return c.formData.Country
	}

	// address line 2
	if AddressLine2Re.MatchString(id) || AddressLine2Re.MatchString(label) ||
		AddressLine2LabelRe.MatchString(label) {
		return c.formData.AddressLine2
	}

	// address line Extra
	if AddressLinesExtraRe.MatchString(label) {
		return c.formData.AddressLineExtra
	}

	// address line 1
	if AddressLine1Re.MatchString(id) || AddressLine1Re.MatchString(label) ||
		AddressLine1LabelRe.MatchString(label) {
		return c.formData.Address
	}

	// zip code
	if ZipCodeRe.MatchString(id) || ZipCodeRe.MatchString(label) {
		return c.formData.ZipCode
	}

	// City
	if CityRe.MatchString(id) || CityRe.MatchString(label) {
		return c.formData.City
	}

	// State
	if StateRe.MatchString(id) || StateRe.MatchString(label) {
		return c.formData.StatePrefecture
	}

	// credit card name
	if NameOnCardRe.MatchString(id) || NameOnCardRe.MatchString(label) {
		return c.formData.FullName
	}

	// CC#
	if CardNumberRe.MatchString(id) || CardNumberRe.MatchString(label) {
		return c.formData.CardNumber
	}

	// CC CVC
	if CardCvcRe.MatchString(id) || CardCvcRe.MatchString(label) {
		return c.formData.CardCVC
	}

	// CC expiration month
	if ExpirationMonthRe.MatchString(id) || ExpirationMonthRe.MatchString(label) {
		return c.formData.ExpirationMonth
	}

	// CC expiration 4 digit year
	if ExpirationYearRe.MatchString(id) || ExpirationYearRe.MatchString(label) ||
		ExpirationDate4DigitYearRe.MatchString(id) || ExpirationDate4DigitYearRe.MatchString(label) {
		return c.formData.ExpirationYear
	}

	// CC expiration 2 digit year
	if ExpirationDate2DigitYearRe.MatchString(id) || ExpirationDate2DigitYearRe.MatchString(label) {
		yearLen := len(c.formData.ExpirationYear)
		if yearLen == 2 {
			return c.formData.ExpirationYear
		}
		if yearLen == 4 {
			return c.formData.ExpirationYear[2:]
		}
	}

	// email
	if EmailRe.MatchString(id) || EmailRe.MatchString(label) {
		return c.formData.Email
	}

	// name/username
	if NameIgnoredRe.MatchString(id) || NameIgnoredRe.MatchString(label) {
		return c.formData.UserName
	}

	// full name
	if NameRe.MatchString(id) || NameRe.MatchString(label) {
		return c.formData.FullName
	}

	// first name
	if FirstNameRe.MatchString(id) || FirstNameRe.MatchString(label) {
		return c.formData.FirstName
	}

	// middle initial name
	if MiddleInitialRe.MatchString(id) || MiddleInitialRe.MatchString(label) {
		return c.formData.MiddleInitial
	}

	// middle name
	if MiddleNameRe.MatchString(id) || MiddleNameRe.MatchString(label) {
		return c.formData.MiddleName
	}

	// last name
	if LastNameRe.MatchString(id) || LastNameRe.MatchString(label) {
		return c.formData.LastName
	}

	// phone
	if PhoneRe.MatchString(id) || PhoneRe.MatchString(label) {
		return c.formData.PhoneNumber
	}

	// phone country code
	if CountryCodeRe.MatchString(id) || CountryCodeRe.MatchString(label) {
		return c.formData.CountryCode
	}

	// phone area code
	if AreaCodeRe.MatchString(id) || AreaCodeRe.MatchString(label) {
		return c.formData.AreaCode
	}

	// phone extensions
	if PhoneExtensionRe.MatchString(id) || PhoneExtensionRe.MatchString(label) {
		return c.formData.Extension
	}

	// passport
	if PassportRe.MatchString(id) || PassportRe.MatchString(label) {
		return c.formData.PassportNumber
	}

	// comment title
	if CommentTitleRe.MatchString(id) || CommentTitleRe.MatchString(label) {
		return c.formData.CommentText
	}

	// comment form
	if CommentRe.MatchString(id) || CommentRe.MatchString(label) {
		return c.formData.CommentText
	}

	// network related
	if NetworkMaskRe.MatchString(id) || NetworkMaskRe.MatchString(label) {
		return c.formData.Network
	}

	// IPV4 address
	if IPAddressRe.MatchString(id) || IPAddressRe.MatchString(label) {
		return c.formData.IPV4
	}

	// IPV6 address
	if IPV6AddressRe.MatchString(id) || IPV6AddressRe.MatchString(label) {
		return c.formData.IPV6
	}

	if TravelOriginRe.MatchString(id) || TravelOriginRe.MatchString(label) {
		return c.formData.TravelOrigin
	}

	if TravelDestinationRe.MatchString(id) || TravelDestinationRe.MatchString(label) {
		return c.formData.TravelDestination
	}

	// we tried our best
	return c.formData.Default
}

// CreateFormContext for a form so we can do analysis on it easier
func (c *CrawlerFormHandler) CreateFormContext(form *browserk.HTMLFormElement) *FormContext {
	formContext := NewFormContext(form.GetAttribute("action"))
	// iterate once to create context
	for i, ele := range form.ChildElements {
		switch ele.Type {
		case browserk.LABEL:
			forHash := ""
			// if it's empty let's just use the next input element as
			// this label *should* be for that input
			if attr := ele.GetAttribute("for"); attr == "" {
				in := form.GetNextOf(i, browserk.INPUT)
				if in != nil {
					forHash = string(in.Hash())
				}
			} else {
				childEle := form.GetChildByNameOrID(attr)
				if childEle != nil {
					forHash = string(childEle.Hash())
				}
			}
			if forHash != "" {
				formContext.AddLabel(forHash, ele.InnerText)
			}
		case browserk.INPUT:
			// we don't want to overwrite if <button type="submit"> already exists as
			// that has precedence
			if ele.GetAttribute("type") == "submit" && formContext.Submit == nil {
				formContext.Submit = ele.Hash()
				continue
			}
			// treat lists as a select where we will ArrowDown -> select
			if ele.GetAttribute("list") != "" {
				continue
			}

			if ele.GetAttribute("type") == "radio" || ele.GetAttribute("type") == "checkbox" {
				continue
			}

			formContext.AddInput(string(ele.Hash()), &InputDetails{
				Name:        strings.ToLower(ele.GetAttribute("name")),
				ID:          strings.ToLower(ele.GetAttribute("id")),
				AriaLabel:   strings.ToLower(ele.GetAttribute("aria-label")),
				Type:        strings.ToLower(ele.GetAttribute("type")),
				PlaceHolder: strings.ToLower(ele.GetAttribute("placeholder")),
				Min:         ele.GetAttribute("min"),
				Max:         ele.GetAttribute("max"),
				Multiple:    false,
				Required:    false,
				Step:        ele.GetAttribute("step"),
				Src:         ele.GetAttribute("src"),
				Alt:         ele.GetAttribute("alt"),
				Pattern:     ele.GetAttribute("pattern"),
				Title:       strings.ToLower(ele.GetAttribute("title")),
			})
		case browserk.TEXTAREA:
			formContext.AddInput(string(ele.Hash()), &InputDetails{
				AriaLabel:   ele.GetAttribute("aria-label"),
				Name:        ele.GetAttribute("name"),
				ID:          ele.GetAttribute("id"),
				PlaceHolder: ele.GetAttribute("placeholder"),
				Max:         ele.GetAttribute("maxlength"),
			})
		case browserk.BUTTON:
			if ele.GetAttribute("type") == "submit" {
				formContext.Submit = ele.Hash()
			}
		}
	}
	return formContext
}
