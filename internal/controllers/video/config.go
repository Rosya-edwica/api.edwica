package video

type Youtube struct {
	Data struct {
		Contents []struct {
			Video struct {
				Id    string `json:"videoId"`
				Title struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"title"`
				Image struct {
					Items []struct {
						Url string `json:"url"`
					} `json:"thumbnails"`
				} `json:"thumbnail"`
			} `json:"videoRenderer"`
		} `json:"contents"`
	} `json:"itemSectionRenderer"`
}
