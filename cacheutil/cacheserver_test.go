package cacheutil
import (
  "fmt"
  "testing"

  "math/rand"
  "reflect"
)

func TestNewInputDataV2(t *testing.T){
  var ipData InputDataV2
  ipData=NewInputDataV2()
  if ipData.(InputDataV2)!=InputDataV2{
    t.Error("New InputData returned nil")
  }

}
