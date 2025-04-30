package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Model
type Kategori struct {
	ID     uint
	Nama   string
	Produk []Produk `gorm:"foreignKey:KategoriID"`
}

type Produk struct {
	ID         uint
	Nama       string
	Harga      float64
	KategoriID uint
	Kategori   Kategori
}

// Koneksi Database
func connectDB() (*gorm.DB, error) {
	dsn := "root:@tcp(localhost:3306)/golang?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrasi Database
func migrasiDB() {
	db, err := connectDB()
	if err != nil {
		log.Fatal("Gagal koneksi DB:", err)
	}
	db.AutoMigrate(&Kategori{}, &Produk{})
	log.Println("Migrasi berhasil. Server berjalan di :8080")
}

// Handler Home
func homeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Gagal koneksi ke database", http.StatusInternalServerError)
		return
	}

	var produk []Produk
	var kategori []Kategori
	db.Preload("Kategori").Find(&produk)
	db.Find(&kategori)

	tmpl := template.Must(template.ParseFiles("template/index.html"))
	tmpl.Execute(w, struct {
		Produk   []Produk
		Kategori []Kategori
	}{
		Produk:   produk,
		Kategori: kategori,
	})
}

// Handler Tambah Kategori
func tambahKategoriHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		nama := r.FormValue("nama")
		db, err := connectDB()
		if err != nil {
			http.Error(w, "Gagal koneksi ke database", http.StatusInternalServerError)
			return
		}
		db.Create(&Kategori{Nama: nama})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handler Tambah Produk
func tambahProdukHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Gagal memproses form", http.StatusBadRequest)
			return
		}

		namaArr := r.Form["nama[]"]
		hargaArr := r.Form["harga[]"]
		kategoriIDArr := r.Form["kategori_id[]"]

		db, err := connectDB()
		if err != nil {
			http.Error(w, "Gagal koneksi ke database", http.StatusInternalServerError)
			return
		}

		for i := range namaArr {
			harga, err1 := strconv.ParseFloat(hargaArr[i], 64)
			kategoriID, err2 := strconv.Atoi(kategoriIDArr[i])
			if err1 != nil || err2 != nil {
				continue // skip data yang tidak valid
			}
			produk := Produk{
				Nama:       namaArr[i],
				Harga:      harga,
				KategoriID: uint(kategoriID),
			}
			if res := db.Create(&produk); res.Error != nil {
				log.Println("Gagal simpan produk:", res.Error)
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handler Edit Produk
func editProdukHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id, _ := strconv.Atoi(r.FormValue("id"))
		nama := r.FormValue("nama")
		harga, _ := strconv.ParseFloat(r.FormValue("harga"), 64)
		kategoriID, _ := strconv.Atoi(r.FormValue("kategori_id"))

		db, err := connectDB()
		if err != nil {
			http.Error(w, "Gagal koneksi ke database", http.StatusInternalServerError)
			return
		}
		db.Model(&Produk{}).Where("id = ?", id).Updates(Produk{
			Nama: nama, Harga: harga, KategoriID: uint(kategoriID),
		})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handler Hapus Produk
func hapusProdukHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Gagal koneksi ke database", http.StatusInternalServerError)
		return
	}
	db.Delete(&Produk{}, id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	migrasiDB()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/tambah-kategori", tambahKategoriHandler)
	http.HandleFunc("/tambah-produk", tambahProdukHandler)
	http.HandleFunc("/edit-produk", editProdukHandler)
	http.HandleFunc("/hapus-produk", hapusProdukHandler)

	log.Println("Server berjalan di :8080")
	http.ListenAndServe(":8080", nil)
}
