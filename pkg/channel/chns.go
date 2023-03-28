package channel

type Channel struct {
  Code   string `json:"code"`
  Name   string `json:"name"`
  Url    string `json:"url"`
  Online *bool  `json:"online"`
}

type ErrorMessage struct {
  Message string `json:"message"`
}
