package vacancy

import (
	"github.com/go-faster/errors"
	"github.com/spf13/viper"
)

type Trudvsem struct {
	Results struct {
		Vacancies []struct {
			Vacancy struct {
				Id             string `json:"id"`
				Name           string `json:"job-name"`
				SalaryFrom     int    `json:"salary_min"`
				SalaryTo       int    `json:"salary_max"`
				Currency       string `json:"currency"`
				Url            string `json:"vac_url"`
				DateUpdate     string `json:"creation-date"`
				Specialisation struct {
					Name string `json:"specialisation"`
				} `json:"category"`
				Addressses struct {
					Address []struct {
						Location string `json:"location"`
					} `json:"address"`
				} `json:"addresses"`
			} `json:"vacancy"`
		} `json:"vacancies"`
	} `json:"results"`
}

type Superjob struct {
	Items []struct {
		Id          int    `json:"id"`
		SalaryFrom  int    `json:"payment_from"`
		SalaryTo    int    `json:"payment_to"`
		Currency    string `json:"currency"`
		PublishedAt int64  `json:"date_published"`
		Name        string `json:"profession"`
		Url         string `json:"link"`
		Company     string `json:"firm_name"`
		Experience  struct {
			Id int `json:"id"`
		} `json:"experience"`

		City struct {
			Id   int    `json:"id"`
			Name string `json:"title"`
		} `json:"town"`

		ProfAreas []struct {
			Name            string `json:"title"`
			Specializations []struct {
				Name string `json:"title"`
			} `json:"positions"`
		} `json:"catalogues"`
	} `json:"objects"`
}

type GeekJob struct {
	Items []GeekJobItem `json:"data"`
}

type GeekJobItem struct {
	Id string `json:"id"`
}

type SuperjobToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TTL          int64  `json:"ttl"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type SuperjobHeaders struct {
	UserAgent string
	Token     string
	SecretId  string
	ClientId  string
}

func InitViper() {
	viper.AddConfigPath("config/")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
}

func setNewSuperjobTokenToConfig(token string) error {
	err := viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "reading config file")
	}
	viper.Set("superjob_token", token)
	err = viper.WriteConfig()
	if err != nil {
		return errors.Wrap(err, "updating config file")
	}
	return nil
}

func GetSuperjobHeaders() (*SuperjobHeaders, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}
	return &SuperjobHeaders{
		UserAgent: viper.GetString("user_agent"),
		Token:     viper.GetString("superjob_token"),
		SecretId:  viper.GetString("superjob_secret_id"),
		ClientId:  viper.GetString("superjob_client_id"),
	}, nil
}

func GetMapSuperjobHeaders() (map[string]string, error) {
	headers, err := GetSuperjobHeaders()
	if err != nil {
		return nil, err
	}
	mapHeaders := make(map[string]string)
	mapHeaders["User-Agent"] = headers.UserAgent
	mapHeaders["Authorization"] = "Bearer " + headers.Token
	mapHeaders["X-Api-App-Id"] = headers.SecretId
	return mapHeaders, nil
}
