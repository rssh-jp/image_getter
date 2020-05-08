package imagegetter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImage(t *testing.T) {
	type Typ struct {
		Name    string
		Content string
		Expects []Data
		Depth   int
	}

	data := []Typ{
		Typ{
			Name: "single",
			Content: `
                <HTML>
                  <BODY>
                    <img src="http://test.jp/test.png"></img>
                  </BODY>
                </HTML>
            `,
			Expects: []Data{
				Data{SrcURL: "http://test.jp/test.png"},
			},
		},
		Typ{
			Name: "multi",
			Content: `
                <HTML>
                  <BODY>
                    <img src="http://test.jp/test.png"></img>
                    <img src="http://test.jp/test2.png"></img>
                  </BODY>
                </HTML>
            `,
			Expects: []Data{
				Data{SrcURL: "http://test.jp/test.png"},
				Data{SrcURL: "http://test.jp/test2.png"},
			},
		},
	}

	for _, item := range data {
		t.Run(item.Name, func(t *testing.T) {

			inst := New()

			defer inst.Close()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, item.Content)
			}))

			defer ts.Close()

			for index, _ := range item.Expects {
				item.Expects[index].BaseURL = ts.URL
			}

			go func() {
				err := inst.Execute(ts.URL, item.Depth)
				if err != nil {
					t.Fatal(err)
				}
			}()

			for i := 0; i < len(item.Expects); i++ {
				select {
				case data := <-inst.URL:
					expect := item.Expects[i]
					actual := data
					if expect.BaseURL != actual.BaseURL || expect.SrcURL != actual.SrcURL {
						t.Errorf("Could not match.\nexpect: %v\nactual: %v\n", expect, actual)
					}
				}
			}
		})
	}
}
