package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type server struct {
	db *sql.DB
}

type OrderInfo struct {
	OrderId        int
	CustomerName   string
	CustomerEmail  string
	OrderTimestamp string
	TotalPrice     int
}

func dbConnect() server {
	db, err := sql.Open("sqlite3", "shop.db")
	if err != nil {
		log.Fatal(err)
	}

	s := server{db: db}

	return s
}

func (s *server) selectOrders() []OrderInfo {
	rows, err := s.db.Query("SELECT id, customer_name, customer_email, order_date, total_price FROM orders;")
	if err != nil {
		log.Fatal(err)
	}

	var orders []OrderInfo
	for rows.Next() {
		var order OrderInfo
		err := rows.Scan(&order.OrderId, &order.CustomerName, &order.CustomerEmail, &order.OrderTimestamp, &order.TotalPrice)
		if err != nil {
			log.Fatal("selectOrders", err)
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		log.Fatal("selectOrders2", err)
	}

	return orders
}

func (s *server) selectOrder(id int) OrderInfo {
	rows := s.db.QueryRow("SELECT id, customer_name, customer_email, order_date, total_price FROM orders WHERE id=?;", id)
	var order OrderInfo
	err := rows.Scan(&order.CustomerName, &order.CustomerEmail, &order.OrderTimestamp, &order.TotalPrice)
	if err != nil {
		log.Fatal("selectOrders", err)
	}

	return order
}

func (s *server) allOrdersHandle(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/orders.html")
	if err != nil {
		log.Fatal("allOrdersHandle", err)
	}

	allOrders := s.selectOrders()
	errExecute := t.Execute(w, allOrders)
	if errExecute != nil {
		log.Fatal("allOrdersHandle2", err)
	}
}


func (s *server) updateOrderByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id := r.FormValue("id")
	idInt, err := strconv.Atoi(id)
	customer_name := r.FormValue("name")
	customer_email := r.FormValue("email")
	updateOrder(customer_name, customer_email, idInt, s)
	http.Redirect(w, r, "/orders", http.StatusSeeOther)
}

func (s *server) updateOrderForm(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/updateOrder.html")
	if err != nil {
		log.Fatal("allOrdersHandle", err)
	}

	err = r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id := r.FormValue("id")
	idInt, err := strconv.Atoi(id)
	order := s.selectOrder(idInt)

	t.Execute(w, order)
}

func (s *server) allOrderChangeHandle(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/updateOrders.html")
	if err != nil {
		log.Fatal("allOrdersHandle", err)
	}

	allOrders := s.selectOrders()
	errExecute := t.Execute(w, allOrders)
	if errExecute != nil {
		log.Fatal("allOrdersHandle2", err)
	}
}

func (s *server) deleteOrder(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id := r.FormValue("id")
	idInt, err := strconv.Atoi(id)
	deleteOrder(idInt, s)
	http.Redirect(w, r, "/index.html", http.StatusSeeOther)
}

func (s *server) formHandle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
	log.Fatal(err)
	}
	customer_name := r.FormValue("name")
	customer_email := r.FormValue("email")
    id := createOrder(customer_name, customer_email, s)


zakaz := OrderInfo{
        OrderId:  id,
		CustomerName: customer_name,
		CustomerEmail: customer_email,
}
 outputHTML(w, "./static/formComplete.html", zakaz)
}

func createOrder(customer_name string, customer_email string, s *server) int {
	
	order_date := time.Now().Format("2006-01-02 15:04:05")
res, err := s.db.Exec("INSERT INTO orders(customer_name, customer_email, order_date, total_price) VALUES (?, ?, ?, ?)", customer_name, customer_email, order_date, 0)
if err != nil {
	log.Fatal(err)
}
order_id, err := res.LastInsertId()
if err != nil {
	log.Fatal(err)
}

return int(order_id)
}
func updateOrder(customer_name string, customer_email string, id int, s *server) int{
    res, err := s.db.Exec("UPDATE orders SET customer_name=?, customer_email=? WHERE id=?", customer_name, customer_email, id)
    if err != nil {
        log.Fatal(err)
    }
    order_id, err := res.RowsAffected()
    if err != nil {
        log.Fatal(err)
    }

    return int(order_id)
}


func deleteOrder(id int, s *server) {
	_, err := s.db.Exec("DELETE FROM orders WHERE id=?", id)
	if err != nil {
	log.Fatal(err)
		}
}

func outputHTML(w http.ResponseWriter, filename string, zakaz OrderInfo) {
	t, err := template.ParseFiles(filename)
	if err != nil {
		log.Fatal(err)
	}

	errExecute := t.Execute(w, zakaz)
	if errExecute != nil {
		log.Fatal(err)
	}
}

func main() {
	s := dbConnect()
	defer s.db.Close()
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/form", s.formHandle)
	http.HandleFunc("/orders", s.allOrdersHandle)
	http.HandleFunc("/change", s.allOrderChangeHandle)
http.HandleFunc("/update", s.updateOrderForm)
http.HandleFunc("/delete", s.deleteOrder)
http.HandleFunc("/updateOrderByID", s.updateOrderByID)
	fmt.Println("Server running...")
	http.ListenAndServe(":8080", nil)
}