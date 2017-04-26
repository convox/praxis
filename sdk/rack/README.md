# Praxis SDK for Go

## Usage

```golang
import (
  "log"
  "github.com/convox/praxis/sdk/rack"
)
    
func main() {
  r := rack.NewFromEnv() // reads $RACK_URL
  
  apps, err := r.AppList()
  if err != nil {
    log.Fatal(err)
  }
  
  for _, app := range apps {
    fmt.Println(app.Name)
  }
}
