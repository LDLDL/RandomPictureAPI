package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ReneKroon/ttlcache/v2"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var config Config
var artists = make(Artists, 0, 100)
var pics = make(Pics, 0, 1000)
var artistsInfo = make(ArtistsInfo, 0, 100)
var artistsMap = make(ArtistsMap)
var randPicReqIDCache = ttlcache.NewCache()

type RandPicP struct {
	PicReqID   string      `json:"pic_req_id"`
	ArtistInfo *ArtistInfo `json:"artist_info"`
}

func init() {
	// read configs
	config.SetDefault()
	if c, err := ioutil.ReadFile("./config.toml"); err == nil {
		if err := toml.Unmarshal(c, &config); err != nil {
			log.Printf("Failed loading config: %s, use default settings.\n", err)
			config.SetDefault()
		} else {
			log.Println("Config loaded.")
		}
	} else {
		log.Println("No config file found. Use default config.")
	}

	// read artists
	var artistsConf ArtistsConf
	if artistsJson, err := ioutil.ReadFile("./pics/artists.json"); err == nil {
		if err := json.Unmarshal(artistsJson, &artistsConf); err != nil {
			log.Printf("Failed loading artists: %s, exiting.\n", err)
			os.Exit(1)
		}
		for _, artistsConfItem := range artistsConf {
			// artist with pics list
			artist := &Artist{
				ArtistInfo: &ArtistInfo{
					Name: artistsConfItem.Name,
					PID: artistsConfItem.PID,
				},
			}
			dir := fmt.Sprintf("./pics/%s/", artistsConfItem.Dir)
			// put pics into pics list
			if files, err := os.ReadDir(dir); err != nil {
				log.Printf("Failed reading directory: %s, ignored.\n", err)
			} else {
				for _, file := range files {
					artist.Pics = append(
						artist.Pics,
						fmt.Sprintf("./pics/%s/%s", artistsConfItem.Dir, file.Name()),
					)
				}
				artist.ArtistInfo.Count = len(artist.Pics)
				// put artist into artists(a list of artist)
				artists = append(artists, artist)
			}
		}

		// also, put into pics list and artists map and artist info list
		for _, artist := range artists {
			for _, picDir := range artist.Pics {
				pics = append(pics, Pic{
					PicPath:    picDir,
					ArtistInfo: artist.ArtistInfo,
				})
			}
			artistsInfo = append(artistsInfo, artist.ArtistInfo)
			artistsMap[artist.ArtistInfo.PID] = artist
		}

		log.Println("Artists loaded.")
	} else {
		log.Println("artists.json not found, exiting.")
		os.Exit(1)
	}

	_ = randPicReqIDCache.SetTTL(5 * time.Minute)
	rand.Seed(time.Now().Unix())
}

func verifyToken(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	if authKey != config.Auth.Key {
		c.String(http.StatusUnauthorized, "Invalid signature")
		c.Abort()
		return
	}
	c.Next()
}

func generateRandomPicRequestId() string {
	var buf = make([]byte, 16)
	now := time.Now().Unix()
	binary.BigEndian.PutUint64(buf, uint64(now))
	rand.Read(buf[8:])

	hash := sha1.New()
	hash.Write(buf)

	return hex.EncodeToString(hash.Sum(nil))
}


func handleArtists(c *gin.Context) {
	c.JSON(http.StatusOK, artistsInfo)
}

func handleRandPicArtist(c *gin.Context) {
	var reqID = generateRandomPicRequestId()
	// select a random artist
	var artistIndex = rand.Intn(len(artists))
	var artist = artists[artistIndex]
	// and then select a random picture
	var picIndex = rand.Intn(artist.ArtistInfo.Count)
	var picPath = artist.Pics[picIndex]

	if err := randPicReqIDCache.Set(reqID, picPath); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, RandPicP{
		PicReqID:   reqID,
		ArtistInfo: artist.ArtistInfo,
	})
}

func handleRandPicPics(c *gin.Context) {
	var reqID = generateRandomPicRequestId()
	// select a pic from pics
	var picIndex = rand.Intn(len(pics))
	var pic = pics[picIndex]

	if err := randPicReqIDCache.Set(reqID, pic.PicPath); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, RandPicP{
		PicReqID:   reqID,
		ArtistInfo: pic.ArtistInfo,
	})
}

func handleRandPicPID(c *gin.Context) {
	var reqID = generateRandomPicRequestId()
	// select artist by pid
	var pid = c.Param("pid")
	var artist = artistsMap[pid]
	// and then select a random picture
	var picIndex = rand.Intn(artist.ArtistInfo.Count)
	var picPath = artist.Pics[picIndex]

	if err := randPicReqIDCache.Set(reqID, picPath); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, RandPicP{
		PicReqID:   reqID,
		ArtistInfo: artist.ArtistInfo,
	})
}

func getImageFile(c *gin.Context)  {
	// request id
	var reqID = c.Param("reqid")
	// get image path from ttl cache
	if picPath, err := randPicReqIDCache.Get(reqID); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else {
		// send pic
		if path, ok := picPath.(string); ok {
			c.File(path)
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}

func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	api := router.Group("/api", verifyToken)
	{
		api.GET("/artists", handleArtists)
		api.GET("/random/artists", handleRandPicArtist)
		api.GET("/random/pics", handleRandPicPics)
		api.GET("/random/artists/:pid", handleRandPicPID)
	}
	router.GET("/image/:reqid", getImageFile)
	router.HandleMethodNotAllowed = true

	if config.Site.Cert != "" && config.Site.Key != "" {
		_ = router.RunTLS(config.Site.Listen, config.Site.Cert, config.Site.Key)
	} else {
		_ = router.Run(config.Site.Listen)
	}
}
