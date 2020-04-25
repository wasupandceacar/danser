package utils

import (
	"os"
	_ "image/jpeg"
	_ "golang.org/x/image/bmp"
	_ "image/png"
	"image"
	"image/draw"
	"log"
	"danser/render/texture"
	"sort"
)

func LoadImage(path string) (*image.NRGBA, error) {
	file, err := os.Open(path)
	log.Println("Loading texture: ", path)
	if err != nil {
		log.Println("er1")
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		log.Println("er2")
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

/*func LoadTexture(path string) (*texture.Texture, error) {
	img, err := LoadImage(path)
	if err == nil {
		tex := glhf.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			4,
			true,
			img.Pix,
		)

		tex.Begin()
		tex.SetWrap(glhf.CLAMP_TO_EDGE)
		tex.End()

		return tex, nil
	}
	return nil, err
}*/

func LoadTexture(path string) (*texture.TextureSingle, error) {
	img, err := LoadImage(path)
	if err == nil {
		tex := texture.LoadTextureSingle(img, 4)

		return tex, nil
	}
	return nil, err
}

func LoadTextureToAtlas(atlas *texture.TextureAtlas, path string) (*texture.TextureRegion, error) {
	img, err := LoadImage(path)
	if err == nil {
		return atlas.AddTexture(path, img.Bounds().Dx(), img.Bounds().Dy(), img.Pix), nil
	}
	log.Println(err)
	return nil, err
}

/*func LoadTextureU(path string) (*glhf.Texture, error) {
	img, err := LoadImage(path)
	if err == nil {
		tex := glhf.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			0,
			true,
			img.Pix,
		)

		tex.Begin()
		tex.SetWrap(glhf.CLAMP_TO_EDGE)
		tex.End()

		return tex, nil
	}
	return nil, err
}*/

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 对数组排序，获得其序号（正序）
func SortRankLowToHigh(array []float64) (rank []int) {
	var sortarray []float64
	sortarray = make([]float64, len(array))
	copy(sortarray, array)
	sort.Float64s(sortarray)
	rank = make([]int, len(array))
	for i, ar := range array {
		rank[i] = firstindexof(sortarray, ar) + 1
	}
	return rank
}

// 对数组排序，获得其序号（反序）
func SortRankHighToLow(array []float64) (rank []int) {
	var sortarray []float64
	sortarray = make([]float64, len(array))
	copy(sortarray, array)
	sort.Float64s(sortarray)
	rank = make([]int, len(array))
	for i, ar := range array {
		rank[i] = len(array) - lastindexof(sortarray, ar)
	}
	return rank
}

// 查找第一个指定元素返回的下标
func firstindexof(array []float64, ar float64) int {
	error := 0.1
	for i, a := range array {
		if (ar >= a - error) && (ar <= a + error) {
			return i
		}
	}
	return -1
}

// 查找最后一个指定元素返回的下标
func lastindexof(array []float64, ar float64) int {
	error := 0.1
	for i := len(array)-1; i >= 0; i-- {
		a := array[i]
		if (ar >= a - error) && (ar <= a + error) {
			return i
		}
	}
	return -1
}

