package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	_ "image/jpeg"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// UserConversion tracks conversions per user
var userConversions = make(map[string]int)
var mu sync.Mutex

var hfToken = os.Getenv("HF_TOKEN")

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/convert", convertHandler)
	http.HandleFunc("/sample", sampleHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpl.Execute(w, nil)
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		fmt.Fprintf(w, `{"error": "Method not allowed"}`)
		return
	}

	// Get user ID from cookie
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error": "Failed to get user ID"}`)
		return
	}

	// Check conversion limit
	mu.Lock()
	if userConversions[userID] >= 10 {
		mu.Unlock()
		w.WriteHeader(403)
		fmt.Fprintf(w, `{"error": "Free trial limit reached"}`)
		return
	}
	userConversions[userID]++
	remaining := 10 - userConversions[userID]
	mu.Unlock()

	// Get form values
	category := r.FormValue("category")
	style := r.FormValue("style")
	file, _, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "Failed to read image"}`)
		return
	}
	defer file.Close()

	// Read file into buffer
	imgData, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "Failed to read image data"}`)
		return
	}

	// Validate image format
	_, _, err = image.Decode(bytes.NewReader(imgData))
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "Invalid image format"}`)
		return
	}

	// Process image based on category
	var result []byte
	switch category {
	case "bw":
		result, err = convertToBW(bytes.NewReader(imgData), style)
	case "cartoon":
		result, err = convertToCartoon(bytes.NewReader(imgData), style)
	case "removebg":
		result, err = removeBG(bytes.NewReader(imgData), style)
	case "changebg":
		result, err = changeBG(bytes.NewReader(imgData), style)
	default:
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"error": "Invalid category"}`)
		return
	}
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, `{"error": "Processing failed: %s"}`, err.Error())
		return
	}

	// Set cookie with updated remaining
	setCookie(w, userID, remaining)

	// Return result
	fmt.Fprintf(w, `{"remaining": %d, "image": "data:image/png;base64,%s"}`, remaining, base64.StdEncoding.EncodeToString(result))
}

func getUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("userID")
	if err != nil {
		// Generate new ID
		id := generateID()
		return id, nil
	}
	return cookie.Value, nil
}

func setCookie(w http.ResponseWriter, userID string, remaining int) {
	http.SetCookie(w, &http.Cookie{Name: "userID", Value: userID, Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "remaining", Value: strconv.Itoa(remaining), Path: "/"})
}

func generateID() string {
	return "user123" // Simple for demo
}

func convertToBW(r io.Reader, style string) ([]byte, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	bwImg := image.NewGray(bounds)
	draw.Draw(bwImg, bounds, img, bounds.Min, draw.Src)
	var buf bytes.Buffer
	png.Encode(&buf, bwImg)
	return buf.Bytes(), nil
}

func convertToCartoon(r io.Reader, style string) ([]byte, error) {
	// If we don't have an external model token, do a local cartoon-like effect
	if hfToken == "" {
		img, _, err := image.Decode(r)
		if err != nil {
			return nil, err
		}
		bounds := img.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

		// Choose quantization levels based on style
		levels := 4
		switch style {
		case "classic":
			levels = 3
		case "modern":
			levels = 6
		case "anime":
			levels = 5
		}

		out := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rC, gC, bC, _ := rgba.At(x, y).RGBA()
				rq := quantize(uint8(rC>>8), levels)
				gq := quantize(uint8(gC>>8), levels)
				bq := quantize(uint8(bC>>8), levels)

				// simple edge detection
				edge := false
				if x+1 < bounds.Max.X {
					r2, g2, b2, _ := rgba.At(x+1, y).RGBA()
					if colorDistanceUint32(rC, r2, gC, g2, bC, b2) > 50000 {
						edge = true
					}
				}
				if y+1 < bounds.Max.Y {
					r2, g2, b2, _ := rgba.At(x, y+1).RGBA()
					if colorDistanceUint32(rC, r2, gC, g2, bC, b2) > 50000 {
						edge = true
					}
				}
				if edge {
					out.Set(x, y, color.RGBA{0, 0, 0, 255})
				} else {
					out.Set(x, y, color.RGBA{rq, gq, bq, 255})
				}
			}
		}

		var buf bytes.Buffer
		png.Encode(&buf, out)
		return buf.Bytes(), nil
	}

	// Otherwise call the external HF model
	imgData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(imgData)
	prompt := "cartoon style of the image"
	if style != "" {
		prompt = style + " cartoon style of the image"
	}
	payload := map[string]interface{}{
		"inputs": b64,
		"parameters": map[string]interface{}{
			"prompt":   prompt,
			"strength": 0.8,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/runwayml/stable-diffusion-v1-5", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+hfToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func removeBG(r io.Reader, style string) ([]byte, error) {
	// Local background removal: detect the corner/background color and make nearby colors transparent
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Sample top-left pixel as background color
	br, bgc, bb, _ := rgba.At(bounds.Min.X, bounds.Min.Y).RGBA()

	out := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := rgba.At(x, y).RGBA()
			// if close to background color, make transparent
			if colorDistanceUint32(r1, br, g1, bgc, b1, bb) < 7000 {
				out.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				r8 := uint8(r1 >> 8)
				g8 := uint8(g1 >> 8)
				b8 := uint8(b1 >> 8)
				out.Set(x, y, color.RGBA{r8, g8, b8, 255})
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, out)
	return buf.Bytes(), nil
}

func changeBG(r io.Reader, style string) ([]byte, error) {
	// If no style provided, fallback to BW
	if style == "" {
		return convertToBW(r, style)
	}
	// If we don't have HF token, perform a simple local background replacement
	if hfToken == "" {
		img, _, err := image.Decode(r)
		if err != nil {
			return nil, err
		}
		bounds := img.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

		// create background
		bgImg := image.NewRGBA(bounds)
		switch style {
		case "blue":
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					bgImg.Set(x, y, color.RGBA{50, 130, 200, 255})
				}
			}
		case "red":
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					bgImg.Set(x, y, color.RGBA{200, 80, 80, 255})
				}
			}
		case "beach":
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					bgImg.Set(x, y, color.RGBA{240, 220, 150, 255})
				}
			}
		case "forest":
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					bgImg.Set(x, y, color.RGBA{90, 140, 70, 255})
				}
			}
		default:
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					bgImg.Set(x, y, color.RGBA{200, 200, 200, 255})
				}
			}
		}

		// Remove original background (simple corner-sample method) into mask
		mask := image.NewRGBA(bounds)
		br, bgc, bb, _ := rgba.At(bounds.Min.X, bounds.Min.Y).RGBA()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r1, g1, b1, _ := rgba.At(x, y).RGBA()
				if colorDistanceUint32(r1, br, g1, bgc, b1, bb) < 7000 {
					mask.Set(x, y, color.RGBA{0, 0, 0, 0})
				} else {
					r8 := uint8(r1 >> 8)
					g8 := uint8(g1 >> 8)
					b8 := uint8(b1 >> 8)
					mask.Set(x, y, color.RGBA{r8, g8, b8, 255})
				}
			}
		}

		// Composite onto bg using Over
		draw.Draw(bgImg, bounds, mask, bounds.Min, draw.Over)

		var buf bytes.Buffer
		png.Encode(&buf, bgImg)
		return buf.Bytes(), nil
	}

	// Otherwise call external HF model
	imgData, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(imgData)
	payload := map[string]interface{}{
		"inputs": b64,
		"parameters": map[string]interface{}{
			"prompt":   "change background to " + style,
			"strength": 0.8,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/runwayml/stable-diffusion-v1-5", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+hfToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// helper: quantize a color into levels
func quantize(c uint8, levels int) uint8 {
	if levels <= 1 {
		return c
	}
	step := 256 / levels
	return uint8((int(c) / step) * step)
}

// helper: squared color distance (operates on 16-bit RGBA values returned by Color.RGBA)
func colorDistanceUint32(r1, r2, g1, g2, b1, b2 uint32) uint32 {
	dr := int32(r1>>8) - int32(r2>>8)
	dg := int32(g1>>8) - int32(g2>>8)
	db := int32(b1>>8) - int32(b2>>8)
	d := int64(dr)*int64(dr) + int64(dg)*int64(dg) + int64(db)*int64(db)
	if d < 0 {
		return ^uint32(0)
	}
	return uint32(d)
}

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	typ := r.URL.Query().Get("type")

	bounds := image.Rect(0, 0, 200, 150)

	var img image.Image

	switch typ {
	case "photo":
		rgbaImg := image.NewRGBA(bounds)
		for y := 0; y < 150; y++ {
			for x := 0; x < 200; x++ {
				rgbaImg.Set(x, y, color.RGBA{uint8(x * 255 / 200), uint8(y * 255 / 150), 128, 255})
			}
		}
		img = rgbaImg
	case "illustration":
		rgbaImg := image.NewRGBA(bounds)
		for y := 0; y < 150; y++ {
			for x := 0; x < 200; x++ {
				rgbaImg.Set(x, y, color.RGBA{255 - uint8(x * 255 / 200), uint8(y * 255 / 150), 200, 255})
			}
		}
		img = rgbaImg
	case "artwork":
		rgbaImg := image.NewRGBA(bounds)
		for y := 0; y < 150; y++ {
			for x := 0; x < 200; x++ {
				rgbaImg.Set(x, y, color.RGBA{128, uint8(x * 255 / 200), uint8(y * 255 / 150), 255})
			}
		}
		img = rgbaImg
	case "before":
		// Choose category-specific realistic-like thumbnails
		cat := r.URL.Query().Get("category")
		switch cat {
		case "bw":
			// Photo-like: muted gradient with a silhouette subject
			rgbaImg := image.NewRGBA(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					v := uint8((x*200+y*100)/350 + 60)
					rgbaImg.Set(x, y, color.RGBA{v, v, v, 255})
				}
			}
			// draw a simple animal silhouette (dog-like)
			for y := 40; y < 120; y++ {
				for x := 40; x < 160; x++ {
					dx := x - 100
					dy := y - 80
					if dx*dx+dy*dy < 1800 {
						rgbaImg.Set(x, y, color.RGBA{30, 30, 30, 255})
					}
					// ear
					if (x-70)*(x-70)+(y-55-60)*(y-55-60) < 250 {
						rgbaImg.Set(x, y, color.RGBA{30, 30, 30, 255})
					}
				}
			}
			img = rgbaImg
		case "cartoon":
			// Cartoon thumbnail: bright background + cartoon face
			rgbaImg := image.NewRGBA(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					rgbaImg.Set(x, y, color.RGBA{uint8(120+(x*90/200)), uint8(80+(y*120/150)), 200, 255})
				}
			}
			// face
			for y := 35; y < 115; y++ {
				for x := 50; x < 150; x++ {
					dx := x - 100
					dy := y - 75
					if dx*dx+dy*dy < 2200 {
						// fill
						rgbaImg.Set(x, y, color.RGBA{255, 210, 160, 255})
						// outline
						if dx*dx+dy*dy > 2000 {
							rgbaImg.Set(x, y, color.RGBA{30, 30, 30, 255})
						}
					}
				}
			}
			// eyes
			for y := 60; y < 68; y++ {
				for x := 82; x < 92; x++ {
					rgbaImg.Set(x, y, color.RGBA{0, 0, 0, 255})
				}
				for x := 108; x < 118; x++ {
					rgbaImg.Set(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
			// smile
			for x := 85; x < 116; x++ {
				y := 92 + int(4*(((x-100)*(x-100))/400))
				if y < 115 {
					rgbaImg.Set(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
			img = rgbaImg
		case "removebg", "changebg":
			// Subject on a natural background (mountain/sky)
			rgbaImg := image.NewRGBA(bounds)
			// background gradient sky -> ground
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					if y < 90 {
						rgbaImg.Set(x, y, color.RGBA{uint8(80 + x*60/200), uint8(150 + y*40/150), uint8(200 + y*20/150), 255})
					} else {
						rgbaImg.Set(x, y, color.RGBA{uint8(80 + x*40/200), uint8(100 + y*20/150), uint8(80 + y*10/150), 255})
					}
				}
			}
			// tree-like subject
			for y := 50; y < 130; y++ {
				for x := 40; x < 120; x++ {
					dx := x - 80
					dy := y - 90
					if dx*dx+dy*dy < 1200 {
						rgbaImg.Set(x, y, color.RGBA{30, 90, 40, 255})
					}
				}
			}
			// trunk
			for y := 85; y < 135; y++ {
				for x := 95; x < 110; x++ {
					rgbaImg.Set(x, y, color.RGBA{100, 60, 40, 255})
				}
			}
			img = rgbaImg
		default:
			rgbaImg := image.NewRGBA(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					rgbaImg.Set(x, y, color.RGBA{200, 200, 200, 255})
				}
			}
			img = rgbaImg
		}

	case "after":
		cat := r.URL.Query().Get("category")
		switch cat {
		case "cartoon":
			rgbaImg := image.NewRGBA(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					r := uint8((x * 255 / 200) ^ (y * 100 / 150))
					g := uint8((y * 255 / 150) ^ (x * 50 / 200))
					b := uint8(180)
					rgbaImg.Set(x, y, color.RGBA{quantize(r, 4), quantize(g, 4), quantize(b, 4), 255})
				}
			}
			img = rgbaImg
		case "removebg":
			rgbaImg := image.NewRGBA(bounds)
			// transparent background
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					rgbaImg.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
			// draw a simple colored oval as subject
			for y := 30; y < 120; y++ {
				for x := 40; x < 160; x++ {
					dx := x - 100
					dy := y - 75
					if dx*dx+dy*dy < 2000 {
						rgbaImg.Set(x, y, color.RGBA{220, 80, 120, 255})
					}
				}
			}
			img = rgbaImg
		case "changebg":
			bg := image.NewRGBA(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					bg.Set(x, y, color.RGBA{90, 140, 70, 255})
				}
			}
			sub := image.NewRGBA(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					sub.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
			for y := 30; y < 120; y++ {
				for x := 40; x < 160; x++ {
					dx := x - 100
					dy := y - 75
					if dx*dx+dy*dy < 2000 {
						sub.Set(x, y, color.RGBA{220, 120, 60, 255})
					}
				}
			}
			draw.Draw(bg, bounds, sub, bounds.Min, draw.Over)
			img = bg
		default:
			grayImg := image.NewGray(bounds)
			for y := 0; y < 150; y++ {
				for x := 0; x < 200; x++ {
					grayImg.SetGray(x, y, color.Gray{uint8((x + y) * 255 / 350)})
				}
			}
			img = grayImg
		}
	default:
		rgbaImg := image.NewRGBA(bounds)
		for y := 0; y < 150; y++ {
			for x := 0; x < 200; x++ {
				rgbaImg.Set(x, y, color.RGBA{200, 200, 200, 255})
			}
		}
		img = rgbaImg
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}