package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	//TODO: should we mock those external services?
	t.Run("External Links", func(t *testing.T) {
		client := http.Client{}

		waitMock := new(waitMock)
		waitMock.On("Wait").Return().Times(3)

		links := []Link{
			Link{
				AbsPath: "https://twitter.com",
				TypeOf:  ExternalLink,
			},
			Link{
				AbsPath: "https://github.com",
				TypeOf:  ExternalLink,
			},
			Link{
				AbsPath: "http://dont.exist.link.com",
				TypeOf:  ExternalLink,
			},
		}

		expected := []Link{
			Link{
				AbsPath: "https://twitter.com",
				TypeOf:  ExternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				AbsPath: "https://github.com",
				TypeOf:  ExternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				AbsPath: "http://dont.exist.link.com",
				TypeOf:  ExternalLink,
				Result: LinkResult{
					Status:  false,
					Message: "404 Not Found",
				},
			},
		}

		valid := NewValidator(client, waitMock)
		result := valid.Links(links)

		assert.Equal(t, expected, result)
	})

	t.Run("Internal Links", func(t *testing.T) {
		links := []Link{
			Link{
				AbsPath: "test-markdowns/external_links.md",
				TypeOf:  InternalLink,
			},
			Link{
				AbsPath: "test-markdowns/sub_path/sub_sub_path/without_links.md",
				TypeOf:  InternalLink,
			},
			Link{
				AbsPath: "test-markdowns/sub_path/absolute_path.md",
				TypeOf:  InternalLink,
			},
			Link{
				AbsPath: "test-markdowns/sub_path/invalid.md",
				TypeOf:  InternalLink,
			},
			Link{
				AbsPath: "test-markdowns/external_links.md#first-header",
				TypeOf:  InternalLink,
			},
			Link{
				AbsPath: "test-markdowns/external_links.md#unknown-header",
				TypeOf:  InternalLink,
			},
		}

		expected := []Link{
			Link{
				AbsPath: "test-markdowns/external_links.md",
				TypeOf:  InternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				AbsPath: "test-markdowns/sub_path/sub_sub_path/without_links.md",
				TypeOf:  InternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				AbsPath: "test-markdowns/sub_path/absolute_path.md",
				TypeOf:  InternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				AbsPath: "test-markdowns/sub_path/invalid.md",
				TypeOf:  InternalLink,
				Result: LinkResult{
					Status:  false,
					Message: "The specified file doesn't exist",
				},
			},
			Link{
				AbsPath: "test-markdowns/external_links.md#first-header",
				TypeOf:  InternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				AbsPath: "test-markdowns/external_links.md#unknown-header",
				TypeOf:  InternalLink,
				Result: LinkResult{
					Status:  false,
					Message: "The specified header doesn't exist in file",
				},
			},
		}

		valid := &Validator{}
		result := valid.Links(links)

		assert.Equal(t, expected, result)
	})

	t.Run("Hash Internal Links", func(t *testing.T) {
		existHeaders := Headers{
			"First Header",
			"Second Header",
			"Third Header",
			"Header with link",
			"Header with block",
			"Very strange header really people create headers look like this",
			"Links",
		}

		links := []Link{
			Link{
				RelPath: "#first-header",
				TypeOf:  HashInternalLink,
			},
			Link{
				RelPath: "#second-header",
				TypeOf:  HashInternalLink,
			},
			Link{
				RelPath: "#third-header",
				TypeOf:  HashInternalLink,
			},
			Link{
				RelPath: "#header",
				TypeOf:  HashInternalLink,
			},
			Link{
				RelPath: "#header-with-block",
				TypeOf:  HashInternalLink,
			},
			Link{
				RelPath: "#header-with-link",
				TypeOf:  HashInternalLink,
			},
			Link{
				RelPath: "#very-strange-header-really-people-create-headers-look-like-this",
				TypeOf:  HashInternalLink,
			},
		}

		expected := []Link{
			Link{
				RelPath: "#first-header",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				RelPath: "#second-header",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				RelPath: "#third-header",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				RelPath: "#header",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status:  false,
					Message: "The specified header doesn't exist in file",
				},
			},
			Link{
				RelPath: "#header-with-block",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				RelPath: "#header-with-link",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
			Link{
				RelPath: "#very-strange-header-really-people-create-headers-look-like-this",
				TypeOf:  HashInternalLink,
				Result: LinkResult{
					Status: true,
				},
			},
		}

		valid := &Validator{}
		result := valid.Links(links, existHeaders)

		assert.Equal(t, expected, result)
	})

	t.Run("Check if throttling works", func(t *testing.T) {
		//GIVEN
		requestRepeats := 5
		client := http.Client{}
		svc := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusTooManyRequests)
		}))

		waitMock := &waitMock{}
		waitMock.On("Wait").Return().Times(requestRepeats)
		v := NewValidator(client, waitMock)
		inputLink := Link{
			TypeOf:  ExternalLink,
			AbsPath: svc.URL,
			Config: &LinkConfig{
				RequestRepeats: &requestRepeats,
			},
		}
		//WHEN
		outLink, err := v.externalLink(inputLink)

		//THEN
		require.NoError(t, err)
		assert.False(t, outLink.Result.Status)
		assert.Equal(t, "Too many requests", outLink.Result.Message)
		waitMock.AssertExpectations(t)
	})
}

type waitMock struct {
	mock.Mock
}

func (m *waitMock) Wait() {
	m.Called()
}
