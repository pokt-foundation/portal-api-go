package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"

	"github.com/pokt-foundation/portal-api-go/relay"
)

const idLength = 24

var (
	appsPath = regexp.MustCompile(`[[:alnum:]]+/v1/([[:alnum:]]+[[:alnum:]-~]*[[:alnum:]]+)$`)
	lbsPath  = regexp.MustCompile(`[[:alnum:]]+/v1/[l|L][b|B]/([[:alnum:]]+[[:alnum:]-~]*[[:alnum:]]+)$`)
	idExp    = regexp.MustCompile(`^[[:alnum:]-]{24}~`)
)

type HttpRequestError error

var (
	ErrInvalidPath HttpRequestError = fmt.Errorf("Path does not match any of the accepted paths.")
)

func ids(path string) (string, string, string, error) {
	match := func(r *regexp.Regexp, p string) string {
		matches := r.FindStringSubmatch(p)
		if len(matches) != 2 {
			return ""
		}
		return matches[1]
	}

	if appID := match(appsPath, path); appID != "" {
		if len(appID) > idLength {
			return appID[:idLength], "", getRelayPath(appID), nil
		}
		return appID, "", "", nil
	}

	if lbID := match(lbsPath, path); lbID != "" {
		if len(lbID) > idLength {
			return "", lbID[:idLength], getRelayPath(lbID), nil
		}
		return "", lbID, "", nil
	}
	return "", "", "", ErrInvalidPath
}

func getRelayPath(id string) string {
	if !idExp.MatchString(id) {
		return ""
	}

	return strings.ReplaceAll(id[idLength:], "~", "/")
}

func buildRelayOptions(req *http.Request) (relay.RelayOptions, error) {
	appID, lbID, relayPath, err := ids(req.URL.Path)
	if err != nil {
		return relay.RelayOptions{}, err
	}

	pathParts := strings.Split(req.URL.Path, ".")
	if len(pathParts) < 1 {
		return relay.RelayOptions{}, fmt.Errorf("%w path: %s", ErrInvalidPath, req.URL.Path)
	}

	relayOptions := relay.RelayOptions{
		ApplicationID:  appID,
		LoadBalancerID: lbID,
		Path:           relayPath,
		Method:         "POST",
		RequestID:      uuid.New(),
		Host:           pathParts[0],
		IP:             req.RemoteAddr,
	}

	origins, ok := req.Header["Origin"]
	if ok {
		relayOptions.Origin = origins[0]
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return relay.RelayOptions{}, fmt.Errorf("Error reading request body: %w", err)
	}

	type requestBody struct {
		BlockchainID string            `json:"blockchainID"`
		RawData      map[string]string `json:"rawData"`
	}

	var reqBody requestBody
	if err := json.Unmarshal(body, &reqBody); err != nil {
		return relay.RelayOptions{}, fmt.Errorf("Error unmarshalling request body: %w", err)
	}
	data, err := parseRawData(reqBody.RawData)
	if err != nil {
		return relay.RelayOptions{}, fmt.Errorf("Error marshalling raw data: %w", err)
	}

	relayOptions.BlockchainID = reqBody.BlockchainID
	relayOptions.RawData = data

	return relayOptions, nil
}

// serves: /v1/{id}, /v1/lb/{id}
func GetHttpServer(r relay.Relayer, l *logger.Logger) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log := l.WithFields(logger.Fields{"Request": *req})
		if req.Method != http.MethodPost {
			log.Warn("Incorrect request method, expected: " + http.MethodPost)
			// fmt.Fprintf(w, "Incorrect request method, expected: %s, got %s", http.MethodPost, req.Method)
			http.Error(w, fmt.Sprintf("Incorrect request method, expected: %s, got: %s", http.MethodPost, req.Method), http.StatusBadRequest)
			return
		}

		relayOptions, err := buildRelayOptions(req)
		if err != nil {
			log.WithFields(logger.Fields{"error": err}).Warn("Failed to build relay request from http request")
			fmt.Fprintf(w, "invalid request")
			return
		}
		log = log.WithFields(logger.Fields{"relayOptions": relayOptions})
		log.Info("Build relay request from http request")

		if relayOptions.LoadBalancerID != "" {
			err = r.RelayWithLb(relayOptions)
		} else {
			err = r.RelayWithApp(relayOptions)
		}
		if err != nil {
			log.WithFields(logger.Fields{"error": err}).Warn("Error relaying")
			// fmt.Fprintf(w, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Info("Relay sent")
		fmt.Fprintf(w, "Relaysent")
	}
}

// TODO: Verify this data type can handle all possible raw data input
func parseRawData(rawData map[string]string) (string, error) {
	b, err := json.Marshal(rawData)
	if err != nil {
		return "", err
	}
	// TODO: if necessary: strings.Replace(options.Data, `\`, "", -1),
	return string(b), nil
}
