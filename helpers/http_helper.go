package helpers

import (
	"math/rand/v2"
	"net/http"
	"strings"
)

var (
	// Generated from https://iplogger.org/useragents/
	USER_AGENTS = []string{
		"windows|Mozilla/5.0 (Windows; Windows NT 6.1; Win64; x64; en-US) Gecko/20130401 Firefox/69.8",
		"explorer|Mozilla/5.0 (compatible; MSIE 11.0; Windows; Windows NT 6.0; x64 Trident/7.0)",
		"linux|Mozilla/5.0 (Linux x86_64; en-US) AppleWebKit/603.49 (KHTML, like Gecko) Chrome/54.0.3195.105 Safari/600",
		"mobile|Mozilla/5.0 (iPod; CPU iPod OS 11_7_3; like Mac OS X) AppleWebKit/600.15 (KHTML, like Gecko)  Chrome/47.0.2333.149 Mobile Safari/534.4",
		"linux|Mozilla/5.0 (U; Linux x86_64) AppleWebKit/600.18 (KHTML, like Gecko) Chrome/52.0.3915.268 Safari/537",
		"windows|Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/600.26 (KHTML, like Gecko) Chrome/50.0.1527.335 Safari/600",
		"explorer|Mozilla/5.0 (compatible; MSIE 10.0; Windows; Windows NT 6.3;; en-US Trident/6.0)",
		"chrome|Mozilla/5.0 (Linux; Linux i661 x86_64) AppleWebKit/601.50 (KHTML, like Gecko) Chrome/53.0.3295.237 Safari/534",
		"explorer|Mozilla/5.0 (compatible; MSIE 10.0; Windows; Windows NT 6.3; Win64; x64; en-US Trident/6.0)",
		"mac|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_7_3) Gecko/20100101 Firefox/58.8",
		"explorer|Mozilla/5.0 (compatible; MSIE 9.0; Windows; U; Windows NT 6.1; Win64; x64 Trident/5.0)",
		"mac|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 7_3_4; en-US) AppleWebKit/533.4 (KHTML, like Gecko) Chrome/53.0.2827.231 Safari/533",
		"firefox|Mozilla/5.0 (Linux i656 x86_64; en-US) Gecko/20130401 Firefox/57.5",
		"windows|Mozilla/5.0 (compatible; MSIE 8.0; Windows; Windows NT 6.1;; en-US Trident/4.0)",
		"linux|Mozilla/5.0 (Linux; U; Linux x86_64; en-US) Gecko/20100101 Firefox/62.8",
		"chrome|Mozilla/5.0 (Windows; Windows NT 6.2; WOW64) AppleWebKit/536.15 (KHTML, like Gecko) Chrome/47.0.3706.174 Safari/602",
		"mac|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 7_0_9) AppleWebKit/537.24 (KHTML, like Gecko) Chrome/54.0.1472.199 Safari/536",
		"firefox|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 7_6_0) Gecko/20130401 Firefox/60.3",
		"linux|Mozilla/5.0 (Linux; U; Linux i683 x86_64) Gecko/20130401 Firefox/70.5",
		"mac|Mozilla/5.0 (Macintosh; Intel Mac OS X 10_5_7) Gecko/20100101 Firefox/61.2",
		"edge|Mozilla/5.0 (Windows NT 10.4; Win64; x64; en-US) AppleWebKit/537.2 (KHTML, like Gecko) Chrome/49.0.3030.117 Safari/533.3 Edge/16.18734",
		"windows|Mozilla/5.0 (compatible; MSIE 7.0; Windows; U; Windows NT 10.4; WOW64; en-US Trident/4.0)",
		"mobile|Mozilla/5.0 (Linux; U; Android 7.1; LG-H910 Build/NRD90M) AppleWebKit/601.1 (KHTML, like Gecko)  Chrome/49.0.2305.281 Mobile Safari/601.1",
		"linux|Mozilla/5.0 (U; Linux x86_64; en-US) AppleWebKit/600.10 (KHTML, like Gecko) Chrome/51.0.3975.222 Safari/600",
		"windows|Mozilla/5.0 (compatible; MSIE 9.0; Windows; U; Windows NT 6.0; x64; en-US Trident/5.0)",
		"windows|Mozilla/5.0 (Windows NT 10.4; Win64; x64) Gecko/20100101 Firefox/49.8",
		"linux|Mozilla/5.0 (Linux; U; Linux i585 ) AppleWebKit/603.28 (KHTML, like Gecko) Chrome/48.0.3012.392 Safari/537",
		"mobile|Mozilla/5.0 (iPhone; CPU iPhone OS 11_7_9; like Mac OS X) AppleWebKit/534.41 (KHTML, like Gecko)  Chrome/47.0.3510.272 Mobile Safari/601.3",
		"mac|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 7_2_9; en-US) Gecko/20100101 Firefox/55.0",
		"chrome|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 8_6_9; en-US) AppleWebKit/533.43 (KHTML, like Gecko) Chrome/53.0.2806.209 Safari/603",
		"chrome|Mozilla/5.0 (Macintosh; Intel Mac OS X 7_5_5) AppleWebKit/600.38 (KHTML, like Gecko) Chrome/47.0.3406.227 Safari/533",
		"linux|Mozilla/5.0 (Linux; U; Linux x86_64; en-US) AppleWebKit/602.41 (KHTML, like Gecko) Chrome/55.0.1163.257 Safari/600",
		"windows|Mozilla/5.0 (Windows; U; Windows NT 6.3;; en-US) AppleWebKit/600.48 (KHTML, like Gecko) Chrome/51.0.1512.224 Safari/537.8 Edge/8.45288",
		"edge|Mozilla/5.0 (Windows NT 10.4; x64) AppleWebKit/602.31 (KHTML, like Gecko) Chrome/54.0.3325.147 Safari/533.5 Edge/15.58691",
		"linux|Mozilla/5.0 (Linux; Linux x86_64) Gecko/20130401 Firefox/59.4",
		"android|Mozilla/5.0 (Android; Android 4.3.1; Ascend G330 Build/JLS36I) AppleWebKit/601.20 (KHTML, like Gecko)  Chrome/50.0.2407.103 Mobile Safari/600.9",
		"firefox|Mozilla/5.0 (Linux; Linux i584 ; en-US) Gecko/20100101 Firefox/71.0",
		"android|Mozilla/5.0 (Linux; Android 5.0.2; Lenovo A7000-a Build/LRX21M;) AppleWebKit/535.8 (KHTML, like Gecko)  Chrome/51.0.3538.238 Mobile Safari/600.8",
		"explorer|Mozilla/5.0 (compatible; MSIE 8.0; Windows; Windows NT 6.2; x64 Trident/4.0)",
		"mobile|Mozilla/5.0 (Android; Android 4.4.1; SM-T531 Build/KOT49H) AppleWebKit/600.18 (KHTML, like Gecko)  Chrome/51.0.3223.366 Mobile Safari/535.6",
		"edge|Mozilla/5.0 (Windows; U; Windows NT 10.4;; en-US) AppleWebKit/534.21 (KHTML, like Gecko) Chrome/49.0.2464.213 Safari/601.8 Edge/15.25436",
		"edge|Mozilla/5.0 (Windows; U; Windows NT 10.4; x64; en-US) AppleWebKit/602.10 (KHTML, like Gecko) Chrome/48.0.1862.122 Safari/600.2 Edge/15.76947",
		"edge|Mozilla/5.0 (Windows; Windows NT 10.3; Win64; x64; en-US) AppleWebKit/534.34 (KHTML, like Gecko) Chrome/48.0.2597.216 Safari/603.9 Edge/18.64807",
		"android|Mozilla/5.0 (Android; Android 4.4; LG-V500 Build/KOT49I) AppleWebKit/600.2 (KHTML, like Gecko)  Chrome/49.0.3200.264 Mobile Safari/601.9",
		"linux|Mozilla/5.0 (Linux; U; Linux x86_64) Gecko/20130401 Firefox/67.9",
		"edge|Mozilla/5.0 (Windows NT 10.1;; en-US) AppleWebKit/536.13 (KHTML, like Gecko) Chrome/54.0.3021.154 Safari/600.8 Edge/16.45119",
		"firefox|Mozilla/5.0 (Linux x86_64; en-US) Gecko/20130401 Firefox/71.9",
		"windows|Mozilla/5.0 (Windows; Windows NT 6.0; WOW64) AppleWebKit/535.36 (KHTML, like Gecko) Chrome/54.0.2451.165 Safari/535",
		"linux|Mozilla/5.0 (U; Linux i575 ; en-US) AppleWebKit/535.46 (KHTML, like Gecko) Chrome/49.0.2322.190 Safari/536",
		"windows|Mozilla/5.0 (Windows; Windows NT 6.2; x64) AppleWebKit/600.18 (KHTML, like Gecko) Chrome/48.0.1625.224 Safari/534",
		"iphone|Mozilla/5.0 (iPhone; CPU iPhone OS 9_7_9; like Mac OS X) AppleWebKit/535.26 (KHTML, like Gecko)  Chrome/48.0.1766.219 Mobile Safari/535.7",
		"explorer|Mozilla/5.0 (compatible; MSIE 10.0; Windows; Windows NT 6.3; WOW64 Trident/6.0)",
		"iphone|Mozilla/5.0 (iPhone; CPU iPhone OS 11_8_2; like Mac OS X) AppleWebKit/534.2 (KHTML, like Gecko)  Chrome/53.0.3944.363 Mobile Safari/603.4",
		"firefox|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_0_7; en-US) Gecko/20130401 Firefox/65.1",
		"edge|Mozilla/5.0 (Windows; U; Windows NT 10.2; Win64; x64) AppleWebKit/537.22 (KHTML, like Gecko) Chrome/51.0.1550.107 Safari/603.8 Edge/15.78338",
		"linux|Mozilla/5.0 (Linux; Linux i543 x86_64; en-US) AppleWebKit/536.29 (KHTML, like Gecko) Chrome/54.0.2687.332 Safari/600",
		"linux|Mozilla/5.0 (Linux; Linux i566 ; en-US) Gecko/20130401 Firefox/46.0",
		"android|Mozilla/5.0 (Linux; Android 5.1.1; SAMSUNG SM-G9258 Build/LMY47X) AppleWebKit/536.42 (KHTML, like Gecko)  Chrome/47.0.2304.292 Mobile Safari/603.3",
		"android|Mozilla/5.0 (Linux; U; Android 5.0.1; SAMSUNG-SM-N915G Build/LRX22C) AppleWebKit/601.26 (KHTML, like Gecko)  Chrome/52.0.2533.227 Mobile Safari/600.5",
		"edge|Mozilla/5.0 (Windows; Windows NT 10.1; WOW64; en-US) AppleWebKit/601.33 (KHTML, like Gecko) Chrome/53.0.2752.223 Safari/600.4 Edge/12.75104",
		"linux|Mozilla/5.0 (U; Linux x86_64; en-US) Gecko/20130401 Firefox/54.9",
		"explorer|Mozilla/5.0 (compatible; MSIE 11.0; Windows NT 6.2; x64; en-US Trident/7.0)",
		"linux|Mozilla/5.0 (Linux i644 x86_64; en-US) Gecko/20100101 Firefox/56.7",
		"windows|Mozilla/5.0 (Windows; Windows NT 10.5; Win64; x64) AppleWebKit/600.39 (KHTML, like Gecko) Chrome/51.0.2933.177 Safari/600.4 Edge/18.58448",
		"windows|Mozilla/5.0 (Windows NT 10.1; x64; en-US) Gecko/20100101 Firefox/63.6",
		"linux|Mozilla/5.0 (Linux; U; Linux x86_64; en-US) AppleWebKit/534.10 (KHTML, like Gecko) Chrome/52.0.2587.368 Safari/535",
		"mobile|Mozilla/5.0 (iPad; CPU iPad OS 9_1_5 like Mac OS X) AppleWebKit/603.44 (KHTML, like Gecko)  Chrome/52.0.1794.266 Mobile Safari/535.9",
		"chrome|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_7) AppleWebKit/535.36 (KHTML, like Gecko) Chrome/53.0.2383.112 Safari/534",
		"edge|Mozilla/5.0 (Windows; Windows NT 10.5;; en-US) AppleWebKit/600.10 (KHTML, like Gecko) Chrome/50.0.1424.279 Safari/533.3 Edge/12.43003",
		"explorer|Mozilla/5.0 (compatible; MSIE 9.0; Windows; Windows NT 6.2; Win64; x64; en-US Trident/5.0)",
		"android|Mozilla/5.0 (Android; Android 4.4.4; Nexus S 4G Build/GRJ22) AppleWebKit/601.32 (KHTML, like Gecko)  Chrome/55.0.3372.188 Mobile Safari/600.8",
		"linux|Mozilla/5.0 (Linux; Linux i576 ; en-US) Gecko/20100101 Firefox/71.6",
		"explorer|Mozilla/5.0 (compatible; MSIE 11.0; Windows; Windows NT 6.1; WOW64; en-US Trident/7.0)",
		"mac|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_7; en-US) Gecko/20100101 Firefox/50.6",
		"mac|Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_9_1; en-US) AppleWebKit/533.50 (KHTML, like Gecko) Chrome/49.0.2959.327 Safari/601",
		"mobile|Mozilla/5.0 (iPhone; CPU iPhone OS 10_9_0; like Mac OS X) AppleWebKit/537.19 (KHTML, like Gecko)  Chrome/49.0.3188.354 Mobile Safari/536.7",
		"edge|Mozilla/5.0 (Windows; Windows NT 10.3;; en-US) AppleWebKit/534.29 (KHTML, like Gecko) Chrome/47.0.1840.346 Safari/601.6 Edge/15.30835",
		"windows|Mozilla/5.0 (Windows; U; Windows NT 6.0; WOW64) AppleWebKit/602.47 (KHTML, like Gecko) Chrome/47.0.3239.387 Safari/537",
		"mobile|Mozilla/5.0 (iPad; CPU iPad OS 11_8_3 like Mac OS X) AppleWebKit/601.23 (KHTML, like Gecko)  Chrome/47.0.1197.338 Mobile Safari/602.4",
		"firefox|Mozilla/5.0 (Linux x86_64; en-US) Gecko/20130401 Firefox/67.2",
		"linux|Mozilla/5.0 (U; Linux i583 x86_64; en-US) AppleWebKit/533.38 (KHTML, like Gecko) Chrome/55.0.3871.310 Safari/602",
		"firefox|Mozilla/5.0 (Linux; Linux x86_64) Gecko/20130401 Firefox/73.8",
		"explorer|Mozilla/5.0 (compatible; MSIE 10.0; Windows; Windows NT 6.3; Win64; x64 Trident/6.0)",
		"mobile|Mozilla/5.0 (iPad; CPU iPad OS 7_8_0 like Mac OS X) AppleWebKit/602.26 (KHTML, like Gecko)  Chrome/55.0.3469.149 Mobile Safari/537.2",
		"firefox|Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2; en-US) Gecko/20100101 Firefox/54.4",
		"windows|Mozilla/5.0 (Windows; Windows NT 6.3; WOW64; en-US) Gecko/20100101 Firefox/71.8",
		"explorer|Mozilla/5.0 (compatible; MSIE 11.0; Windows; Windows NT 6.0; Trident/7.0)",
		"iphone|Mozilla/5.0 (iPhone; CPU iPhone OS 7_7_0; like Mac OS X) AppleWebKit/600.48 (KHTML, like Gecko)  Chrome/50.0.2027.168 Mobile Safari/533.1",
		"chrome|Mozilla/5.0 (Macintosh; Intel Mac OS X 10_5_8) AppleWebKit/600.37 (KHTML, like Gecko) Chrome/49.0.2255.280 Safari/536",
		"windows|Mozilla/5.0 (Windows; Windows NT 6.3; WOW64; en-US) AppleWebKit/602.8 (KHTML, like Gecko) Chrome/48.0.2973.278 Safari/601.7 Edge/9.82629",
		"firefox|Mozilla/5.0 (Windows; Windows NT 10.1; WOW64; en-US) Gecko/20100101 Firefox/46.8",
		"iphone|Mozilla/5.0 (iPhone; CPU iPhone OS 10_4_4; like Mac OS X) AppleWebKit/534.19 (KHTML, like Gecko)  Chrome/48.0.3585.146 Mobile Safari/603.4",
		"edge|Mozilla/5.0 (Windows; U; Windows NT 10.1; Win64; x64; en-US) AppleWebKit/603.5 (KHTML, like Gecko) Chrome/54.0.2772.289 Safari/533.9 Edge/8.45938",
		"chrome|Mozilla/5.0 (Windows; U; Windows NT 6.0;) AppleWebKit/602.17 (KHTML, like Gecko) Chrome/48.0.2927.207 Safari/600",
		"explorer|Mozilla/5.0 (compatible; MSIE 10.0; Windows; U; Windows NT 6.1; x64; en-US Trident/6.0)",
		"iphone|Mozilla/5.0 (iPhone; CPU iPhone OS 7_5_6; like Mac OS X) AppleWebKit/600.11 (KHTML, like Gecko)  Chrome/52.0.2649.274 Mobile Safari/601.6",
		"android|Mozilla/5.0 (Linux; U; Android 6.0.1; HTC One0P8B2 Build/MRA58K) AppleWebKit/602.47 (KHTML, like Gecko)  Chrome/48.0.2046.198 Mobile Safari/533.5",
		"edge|Mozilla/5.0 (Windows; Windows NT 10.2; x64) AppleWebKit/600.12 (KHTML, like Gecko) Chrome/55.0.3762.379 Safari/533.8 Edge/15.96417",
		"edge|Mozilla/5.0 (Windows; U; Windows NT 10.4; Win64; x64) AppleWebKit/603.5 (KHTML, like Gecko) Chrome/48.0.1491.350 Safari/601.0 Edge/12.36345",
		"windows|Mozilla/5.0 (Windows; U; Windows NT 6.0; WOW64) AppleWebKit/536.20 (KHTML, like Gecko) Chrome/47.0.3369.375 Safari/536",
	}
)

func getUserAgent() string {
	return strings.Split(USER_AGENTS[rand.IntN(len(USER_AGENTS))], "|")[1]
}

func HumanlikeGet(url string) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", getUserAgent())
	return client.Do(req)
}
