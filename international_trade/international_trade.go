package main

import (
    "io/ioutil"
    "encoding/xml"
    "log"
    "os"
    "encoding/csv"
    "strings"
    "strconv"
    "fmt"
)


type Rate struct {
    From string `xml:"from"`
    To string `xml:"to"`
    Conversion float32 `xml:"conversion"`
}

type RatesContainer struct {
    XMLName xml.Name `xml:"rates"`
    Rates []Rate `xml:"rate"`
}

type USDConverter map[string]float32

func GetConversions(filename string) (RatesContainer, error) {
    r := RatesContainer{}

    file_contents, err := ioutil.ReadFile(filename)
    if err != nil {
        return RatesContainer{}, err
    }

    err = xml.Unmarshal(file_contents, &r)
    if err != nil {
        return RatesContainer{}, err
    }

    return r, nil
}

func (u *USDConverter) Populate(r_list []Rate) {
    var failed_rates []Rate
    
    for _, value := range r_list {
        if converted := (*u).addConversion(value); !converted {
            failed_rates = append(failed_rates, value)
        }
    }

    if (len(failed_rates) > 0) && (len(failed_rates) < len(r_list)) {
        // we are still making progress
        u.Populate(failed_rates)
    }

    return
}

// Returns false if the conversion could not be added due to
// neither of the countries being known
func (u *USDConverter) addConversion(r Rate) (bool) {
    switch {
    case r.To == "USD":
        (*u)[r.From] = r.Conversion
        return true
    case r.From == "USD":
        (*u)[r.To] = (1 / r.Conversion)
        return true
    }

    for k, _ := range (*u) {
        switch {
        case r.To == k:
            (*u)[r.From] = ((*u)[r.To]) * r.Conversion
            return true
        case r.From == k:
            (*u)[r.To] = ((*u)[r.From]) * (1 / r.Conversion)
            return true
        }
    }

    return false
}

func GetTotalSales(filename string, item string, converter USDConverter) (int, int, error) {
    file, err := os.Open(filename) // For read access.
    if err != nil {
        return 0, 0, err
    }
    
    csv_r := csv.NewReader(file)
    records, err := csv_r.ReadAll()
    if err != nil {
        return 0, 0, err
    }

    sales := make([][]string, 0)
    for _, record := range records {
        if record[1] == item {
            sales = append(sales, strings.Split(record[2], " "))
        }
    }
    log.Println(sales)

    total := int(0)
    for _, sale := range sales {
        f, err := strconv.ParseFloat(sale[0], 32)
        if err != nil {
            return 0, 0, err
        }
        if sale[1] == "USD" {
            total += int(float32(f) * 100)
        } else {
            total += int((float32(f) * converter[sale[1]]) * 100)
        }
    }

    dollar_total := total / 100
    cents_total := total % 100
    return dollar_total, cents_total, nil
}

func main() {
    r, err := GetConversions("RATES.xml")
    log.Println(r)
    if err != nil {
        log.Fatal(err)
    }
    
    usd := make(USDConverter)
    usd.Populate(r.Rates)
    log.Println(usd)

    dollars, cents, err := GetTotalSales("TRANS.csv", "DM1182", usd)
    if err != nil {
        log.Fatal(err)
    }
    result := fmt.Sprintf("%d.%d\n", dollars, cents)
    log.Println(result)
}