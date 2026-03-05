package handlers

type ProfileResponse struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthPlace string `json:"birth_place"`
	BirthDate  string `json:"birth_date"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	NationalID string `json:"national_id"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Photo      string `json:"photo,omitempty"`
}

type CreateProfileRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BirthPlace string `json:"birth_place"`
	BirthDate  string `json:"birth_date"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	NationalID string `json:"national_id"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Photo      string `json:"photo"`
}

