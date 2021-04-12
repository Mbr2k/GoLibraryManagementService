package main

import (
	// Standard packages:
	"time"
	"fmt"
	"log"
	"os"
	// DB requirements:
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rubenv/sql-migrate"
	// Server requirements:
	"net/http"
	"encoding/json"
	// Crypto requirements:
	"golang.org/x/crypto/bcrypt"
)

// Global Variables:
var db *sql.DB

func main() {

	db = InitialiseDB()
	
	RunMigrations()

	go ReservationExpiryAlarm()
	// Server Config:
	s := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/signup", SignUp)
	http.HandleFunc("/loan", LoanBook)
	http.HandleFunc("/addbook", AddBook)
	http.HandleFunc("/returnbook", ReturnBook)
	http.HandleFunc("/updaterole", UpdateRole)

	log.Fatal(s.ListenAndServe())
}

//---------------------
// Types
//---------------------
type User struct {
	Password 		string 		`json:"password" db:"password"`
	Name 			string 		`json:"name" db:"name"`
	Role			string 		`json:"role" db:"role"`
	ID              uint64		`json:"id" db:"id"`
}	

type Book struct {
	Title 			string 		`json:"title" db:"title"`
	Author			string 		`json:"author" db:"author"`
	NumAvailable	uint64 		`json:"num_available" db:"num_available"`
	NumLoaned		uint64 		`json:"num_loaned" db:"num_loaned"`
}

type Reservation struct {
	Title 			string 		`json:"title" db:"title"`
	Borrower 		string 		`json:"username" db:"username"`
}

type LoanRequest struct {
	Librarian 	User 			`json:"user"`
	Loan 		Reservation		`json:"reservation"`
}

type ReturnRequest struct {
	Librarian 	User 			`json:"user"`
	Return 		Reservation		`json:"reservation"`
}

type AddBookRequest struct {
	Librarian 		User 			`json:"user"`
	Title 			string 			`json:"title" db:"title"`
	Author			string 			`json:"author" db:"author"`
	NumAvailable	uint64 			`json:"num_available" db:"num_available"`
}

type AddRoleRequest struct {
	Issuer 			User 			`json:"issuer"`
	Receiver 		User 			`json:"receiver"`
}

//---------------------
// Type Methods
//---------------------

func (user *User)Authenticate()(error){
	var err error
	// if the user knows the libraryRootPassword then they can bypass normal authentication:
	// - a security conscious admin will delete this environment variable once some senior librarians are on the system
	masterkey , set := os.LookupEnv("libraryRootPassword")
	if(user.Password == masterkey && set){
		return err
	}
	// Get the existing entry present in the database for the given username
	result := db.QueryRow("select password, id from users where name=?", user.Name)

	var storedPw string
	
	// Also scan in our user's ID for later processing
	err = result.Scan(&storedPw, &user.ID)
	fmt.Println(storedPw)
	fmt.Println(err)
	if err != nil {
		return err
	}

	// Compare the stored hashed password, with the hashed version of the password that was received
	err = bcrypt.CompareHashAndPassword([]byte(storedPw), []byte(user.Password))
	// Nil return here means we have authenticated the user successfully
	return err
}

func (user *User)GetRole()(string){
	// Get the existing entry present in the database for the given username
	db.QueryRow("select role from users where name = ?", user.Name).Scan(&user.Role)

	// Nil return here means we have authenticated the user successfully
	return user.Role
}

//---------------------
// Handler Functions
//---------------------

func SignUp(w http.ResponseWriter, r *http.Request)(){
	// Parse and decode the request body into a new `User` instance
	user 	:= &User{}

	err 	:= json.NewDecoder(r.Body).Decode(user)

	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Salt and hash the password using the bcrypt algorithm
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)

	// Next, insert the username, along with the hashed password into the database
	if _, err = db.Query("insert into users (name , password) values (?, ?)", user.Name, string(hashedPassword)); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	// We reach this point if the credentials we correctly stored in the database, and the default status of 200 is sent back
	return

}



func UpdateRole(w http.ResponseWriter, r *http.Request)(){
	roleRequest := &AddRoleRequest{}
	
	// #Begin authentication#
	err := json.NewDecoder(r.Body).Decode(roleRequest)
	
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Now we authenticate the librarian taking this request
	err = roleRequest.Issuer.Authenticate()
	if(err!=nil){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Now check if the issuer has the appropriate role, or if they have the root pw
	role 			:= roleRequest.Issuer.GetRole()
	masterkey , set := os.LookupEnv("libraryRootPassword")
	if(role != "SeniorLibrarian" && ((masterkey == roleRequest.Issuer.Password) && set)){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// #End authentication#

	// Update the receiver user with the role provided against their user
	if _, err = db.Query("UPDATE users SET role = ? WHERE name = ?", roleRequest.Receiver.Role, roleRequest.Receiver.Name); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
}

func AddBook(w http.ResponseWriter, r *http.Request)(){
	// Add book will either add a single new book or several depending on whether numAvailable was specified on the request body or not.
	bookRequest := &AddBookRequest{}
	// #Begin authentication#
	err := json.NewDecoder(r.Body).Decode(bookRequest)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Now we authenticate the librarian taking this request
	err = bookRequest.Librarian.Authenticate()
	if(err!=nil){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	role := bookRequest.Librarian.GetRole()
	fmt.Println(bookRequest.Librarian.Name,":",role)
	// Only Senior Librarians should be able to add books:
	if(role != "SeniorLibrarian"){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// #End authentication#

	// NumAvailable will be parsed to a zero value if not specified and will default to 1, else we specify it on insert:
	if(bookRequest.NumAvailable == 0){
		if _, err = db.Query("insert into books (userCreated_id , title, author) values (?, ?, ?)", bookRequest.Librarian.ID, bookRequest.Title, bookRequest.Author); err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
	}else{
		if _, err = db.Query("insert into books (userCreated_id , title, author, num_available) values (?, ?, ?, ?)", 
		bookRequest.Librarian.ID, bookRequest.Title, bookRequest.Author, bookRequest.NumAvailable); err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
	}
}

func LoanBook(w http.ResponseWriter, r *http.Request)(){
	// AKA Checkout book
	loanRequest := &LoanRequest{}
	// #Begin authentication#
	err := json.NewDecoder(r.Body).Decode(loanRequest)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Now we authenticate the librarian taking this request
	err = loanRequest.Librarian.Authenticate()
	if(err!=nil){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	role := loanRequest.Librarian.GetRole()
	if(role != "Librarian" && role != "SeniorLibrarian"){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// #End authentication#
	
	// Check there aren't too many reservations for the users requesting the reservations:
	var numReservations uint64
	db.QueryRow("SELECT COUNT(id) FROM reservations WHERE username = ?", loanRequest.Loan.Borrower).Scan(&numReservations)
	fmt.Println("# books out now:", numReservations)
	if(numReservations >= 4){
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Too many books currently reserved by this user: %d", numReservations)
		return
	}
	// Check num_available & num_loaned are acceptable to update
	// Check there aren't too many reservations for the users requesting the reservations:
	var numAvailable , numLoaned uint64
	db.QueryRow("SELECT num_available , num_loaned FROM books WHERE title = ?", loanRequest.Loan.Title).Scan(&numAvailable , &numLoaned)
	fmt.Printf("\n# %s currently on shelves: %d\n", loanRequest.Loan.Title , numAvailable)
	if(numAvailable <= 0){
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "No books currently available for this title, however there are %d books out on loan for this title.", numLoaned)
		return
	}

	// Write a transaction to make a reservation and update the num_loaned and num_available fields on books
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function.

	stmt, err := tx.Prepare("INSERT INTO reservations (username , title) VALUES (?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	defer stmt.Close() // Prepared statements take up server resources and should be closed after use.
	if _, err := stmt.Exec(loanRequest.Loan.Borrower, loanRequest.Loan.Title); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	stmt2, err := tx.Prepare("UPDATE books SET num_available = num_available - 1 , num_loaned = num_loaned + 1 WHERE title = ?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	defer stmt2.Close()
	if _, err := stmt2.Exec(loanRequest.Loan.Title); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
}

func ReturnBook(w http.ResponseWriter, r *http.Request)(){
	// AKA Checkin book
	loanRequest := &LoanRequest{}
	// #Begin authentication#
	err := json.NewDecoder(r.Body).Decode(loanRequest)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Now we authenticate the librarian taking this request
	err = loanRequest.Librarian.Authenticate()
	if(err!=nil){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	role := loanRequest.Librarian.GetRole()
	if(role != "Librarian" && role != "SeniorLibrarian"){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// #End authentication#

	// Update the returned books availability
	if _, err = db.Query("UPDATE books SET num_available = num_available + 1 , num_loaned = num_loaned - 1 WHERE title = ?", loanRequest.Loan.Title); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	// Delete this reservation from the database:
	if _, err = db.Query("DELETE FROM reservations WHERE username = ? AND title = ?", loanRequest.Loan.Borrower, loanRequest.Loan.Title); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	return
}

//---------------------
// Reservation Alarm 
//---------------------

func ReservationExpiryAlarm() {

	var username , title string
	var dateTime time.Time
	rows , err := db.Query("SELECT username , title, dateCreated FROM reservations WHERE DATEDIFF(CURDATE(),dateCreated) > 7")

	if(err!=nil){
		log.Println(err.Error())
	}

	for rows.Next() {
		err := rows.Scan(&username, &title , &dateTime)
		if(err!=nil){
			log.Println(err.Error())
		}
		// now we do something with our results if there are any, send an email for alarm for example.
		// I will just print those reservations which are >1week old...
		fmt.Printf("\nThere is a reservation for user %s, for title %s that has been out since %s\n", username, title, dateTime.String())
	}

	time.Sleep(24 * time.Hour)
	
}

//---------------------
// DB Utility Functions
//---------------------
func InitialiseDB() *sql.DB {
	// Returns a pointer to the sql object for the database
	// we wish to connect to at the below address.
	db_url	:= fmt.Sprintf(os.Getenv("DB_USER")+":"+os.Getenv("DB_PW")+"@tcp("+os.Getenv("DB_URL")+")/"+os.Getenv("DB_NAME"))
	db_url 	 = db_url + "?parseTime=true"

	fmt.Println("Accessing Database with identity: " + db_url)

	db, err := sql.Open("mysql",db_url)
	if err != nil {
		fmt.Println("Error validating driver and connection information")
		log.Fatal(err)
	}
	
	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Println(err.Error()) // error logging if db unresponsive
	}

	// DB Config:
	db.SetConnMaxLifetime(10*time.Second)
	db.SetMaxIdleConns(0)

	// Returning pointer to sql object representing database:
	return db
}

func RunMigrations() {
	// Runs the database migration files specified in /db/migrations:
	migrations := &migrate.FileMigrationSource{
		Dir: "db/migrations",
	}

	n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		log.Fatalf("Error while running migrations!; \n%s" , err.Error())
	}

	fmt.Printf("Applied %d migrations!\n", n)
}

