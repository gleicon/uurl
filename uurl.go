package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fiorix/go-redis/redis"
	"github.com/pilu/go-base62"
)

type UURL struct {
	db *redis.Client
}

type URLStats struct {
	Url           string
	Date_creation string
	Referers      []string
	Visitors      []string
	Stats         map[string]string
}

func (us *URLStats) toJson() ([]byte, error) {
	j, err := json.Marshal(us)
	return j, err
}

const (
	UNIQUE_COUNTER         = "URL:UUID"
	URL_MASK               = "URL:%s"
	ENCODED_URL_REF_LIST   = "URL:UID:REFERERS:%s"
	ENCODED_URL_VIS_LIST   = "URL:UID:VISITORS:%s"
	ENCODED_URL_DICT       = "URL:UID:DATA:%s"
	ENCODED_URL_STATS_DICT = "URL:UID:STATS:%s"
	ENCODED_URL_MASK       = "URL:UID:MASK:%s"
)

func (uu *UURL) UpdateURLData(url string, customURL string) (string, error) {
	k := fmt.Sprintf(URL_MASK, url)
	u, err := uu.db.Get(k)
	if err != nil {
		return "", err
	}
	if u == "" {
		var uid string
		if customURL == "" {
			uuid, err := uu.GetUUID()
			if err != nil {
				return "", err
			}
			uid = base62.Encode(uuid)
		} else {
			uid = customURL
		}

		err = uu.db.Set(k, uid)
		if err != nil {
			return "", err
		}

		k = fmt.Sprintf(ENCODED_URL_MASK, uid)
		err = uu.db.Set(k, url)
		if err != nil {
			return "", err
		}
		k = fmt.Sprintf(ENCODED_URL_DICT, uid)
		t := time.Now().Unix()
		xx := strconv.Itoa(int(t))
		err = uu.db.HSet(k, "date_creation", xx)
		if err != nil {
			return "", err
		}
		err = uu.db.HSet(k, "url", url)
		if err != nil {
			return "", err
		}
		return uid, nil
	}
	return u, nil
}

func (uu *UURL) GetUUID() (int, error) {
	u, e := uu.db.Incr(UNIQUE_COUNTER)
	return u, e
}

func (uu *UURL) GetURLByUID(uid, ip, ref string) (string, error) {
	var err error
	var url string

	k := fmt.Sprintf(ENCODED_URL_MASK, uid)

	if url, err = uu.db.Get(k); err != nil {
		return "", err
	}
	if err = uu.UpdateEncodedURLData(uid, ip, ref); err != nil {
		return "", err
	}
	return url, nil
}

func (uu *UURL) UpdateEncodedURLData(uid, ip, ref string) error {
	k := fmt.Sprintf(ENCODED_URL_DICT, uid)
	_, err := uu.db.HIncrBy(k, "clicks", 1)
	if err != nil {
		return err
	}
	t := time.Now()
	dt := fmt.Sprintf("%d.%d.%d.%d", t.Year(), t.Month(), t.Day(), t.Hour())
	ks := fmt.Sprintf(ENCODED_URL_STATS_DICT, uid)
	_, err = uu.db.HIncrBy(ks, dt, 1)
	if err != nil {
		return err
	}

	if ref != "" {
		ee := fmt.Sprintf(ENCODED_URL_REF_LIST, uid)
		_, err = uu.db.LPush(ee, ref)
		if err != nil {
			return err
		}
	}
	if ip != "" {
		ee := fmt.Sprintf(ENCODED_URL_VIS_LIST, uid)
		_, err = uu.db.LPush(ee, ip)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uu *UURL) GetURLStatsByUID(uid string) (*URLStats, error) {
	k := fmt.Sprintf(ENCODED_URL_DICT, uid)
	st, err := uu.db.HGetAll(k)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, nil
	}

	k = fmt.Sprintf(ENCODED_URL_STATS_DICT, uid)
	ss, err := uu.db.HGetAll(k)
	if err != nil {
		return nil, err
	}
	k = fmt.Sprintf(ENCODED_URL_REF_LIST, uid)
	rl, err := uu.db.LRange(k, 0, -1)
	if err != nil {
		return nil, err
	}

	k = fmt.Sprintf(ENCODED_URL_VIS_LIST, uid)
	vl, err := uu.db.LRange(k, 0, -1)
	if err != nil {
		return nil, err
	}

	us := URLStats{Visitors: vl, Referers: rl, Date_creation: st["date_creation"], Url: st["url"], Stats: ss}
	return &us, nil
}

func NewUURL(db *redis.Client) *UURL {
	uu := UURL{db: db}
	return &uu
}
