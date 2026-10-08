package feeds

import "testing"

func TestFeedItemViewDisplayURL(t *testing.T) {
	cases := map[string]string{
		"https://www.example.com/blog/post/": "example.com/blog/post",
		"https://example.com/":               "example.com",
		"https://example.com":                "example.com",
		"http://sub.example.com/a?b=c#d":     "sub.example.com/a",
		"https://example.com/a/b/c/d":        "example.com/a/b/c/d",
		"not a url":                          "not a url",
		"":                                   "",
	}

	for in, want := range cases {
		if got := (FeedItemView{URL: in}).DisplayURL(); got != want {
			t.Errorf("DisplayURL(%q) = %q, want %q", in, got, want)
		}
	}
}
