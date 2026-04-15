
# RESTful Go API for OpenPowerlifting Data 
A RESTful API for open-source powerlifting data.

## Getting Started

### Prerequisites
- Go 1.23+
- Git (to clone the OPL dataset)

### 1. Clone this repo
```bash
git clone https://github.com/Kevin-Aguirre/powerlifting-api.git
cd powerlifting-api
```

### 2. Clone the OpenPowerlifting data repo
The API reads meet data from the [OPL dataset](https://gitlab.com/openpowerlifting/opl-data). Clone it inside the project directory:

```bash
git clone https://gitlab.com/openpowerlifting/opl-data.git
```

This will create an `opl-data/` folder containing the `meet-data/` directory the API needs.

> **Note:** The OPL dataset is large (~1GB+). The clone may take several minutes.

### 3. Update the data path in `main.go`
Open `main.go` and update `dataFolderPath` to point to the cloned data:

```go
var dataFolderPath = "./opl-data/meet-data"
```

### 4. Install dependencies
```bash
go mod download
```

### 5. Run the server
```bash
go run main.go
```

The server starts on **`:8080`**. Once loaded, you can hit endpoints like:

```
GET http://localhost:8080/lifters
GET http://localhost:8080/lifters/{lifterName}
GET http://localhost:8080/lifters/names
GET http://localhost:8080/meets
GET http://localhost:8080/meets/{federationName}
GET http://localhost:8080/federations
```

> **Note:** Loading the full dataset takes a while on startup — this is expected.

---

## TODO
* Include API Documentation guide in this readme or another document linked here 
* Add Filtering and Sorting and unit (lbs or kg) Fields in Requests (especially for all lifters endpoint)
* Expand database to include meet data (not just lifter data)
* Include OPL image here 
* Inlucde Lifter Instagrams and other fields to lifter type (sex, etc. )
* Include Lifter Personal Records 
* Check OPL website for any functionality not listed here
* Break up different packages further for cleaner code
* Add more routes (top-n-lifters, etc.)

## Contributing
If there's a feature missing from this API you'd like to see (that isn't in the TODO list above), feel free to make a pull request, issue, or reach out :) 

## Dependencies
Built with Go version 1.24.4

## OPL Links
* [Open Powerlifting Rankings](https://www.openpowerlifting.org/)
* [Open Powelifting Site](https://openpowerlifting.gitlab.io/opl-csv/)
* [Open Powerlifting Data](https://gitlab.com/openpowerlifting/opl-data)


