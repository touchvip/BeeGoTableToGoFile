package main

import (
	//"bytes"
	"fmt"
	"log"
	"os"
    "os/exec"
    "path/filepath"
	"strings"
	"text/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	//"database/sql/driver"	
	//"net/url"
)


type SqlStruct struct{
	GoFile string
	sqldb *sql.DB
	Tables []Table
}


type Table struct{
	Table_name string
	Columns []Column
}

type Column struct {
	Name       string
	Type       string
	IsPrimary  bool
	IsForeign  bool
	ForeignKey string
}

func NewSqlStruct(goFileName string) (*SqlStruct,error) {
    db, err := sql.Open("mysql", "root:root@/ihome")
    if err != nil {
        log.Fatalf("Open database error: %s\n", err)
    }
    //defer db.Close()	
	err = db.Ping()
    if err != nil {
        log.Fatal(err)
		return nil,nil  
    } else{
	return &SqlStruct{GoFile:goFileName,sqldb:db},nil
	}
	
}
func (slt *SqlStruct) BuildTableStruct()  {
    rows, err := slt.sqldb.Query(`select table_name from information_schema.tables where table_type="base table"`)
    if err != nil {
        log.Println(err)
    } 
    defer rows.Close()

    var Table_Name string
    for rows.Next() {
        err := rows.Scan(&Table_Name)
        if err != nil {
            log.Fatal(err)
        }
		slt.createTableColumns(Table_Name)
        log.Println(Table_Name)
    }	
}
func (slt *SqlStruct) mysqlToGo(col_type string) string  {
	  stype:=strings.ToUpper(col_type)
	  if stype == "VARCHAR" { return "string"}
	  if stype == "CHAR" { return "string"}
	  if stype == "TEXT" { return "string"}
	  if stype == "INT" { return "int64"}
	  if stype == "FLOAT" { return "float32"}
	  if stype == "TINYINT" { return ",int16"}
	  if stype == "SMALLINT" { return ",int16"}
	
	  if stype == "TIMESTAMP" { return "time.Time"}
	  if stype == "DATETIME" { return "time.Time"}
	  if stype == "BOOL" { return "bool"}
	  return ""
}
func (slt *SqlStruct) createTableColumns(table_name string)  {
	table:=Table{Table_name:table_name}
    rows, err := slt.sqldb.Query(`select Column_Name,DATA_TYPE,COLUMN_KEY,EXTRA from 
	 information_schema.columns where  table_name=?`,table_name)
	
	table.Table_name=strings.ToUpper(table.Table_name[0:1])+strings.ToLower(table.Table_name[1:len(table.Table_name)])
    if err != nil {
        log.Println(err)
    } 
    defer rows.Close()
    
    var Column_Name,DATA_TYPE,COLUMN_KEY,EXTRA string
    for rows.Next() {
        err := rows.Scan(&Column_Name,&DATA_TYPE,&COLUMN_KEY,&EXTRA)
        if err != nil {
            log.Fatal(err)
        }
		col:=Column{}
		s:=Column_Name
		EXTRA=strings.ToUpper(EXTRA)
		col.Name=strings.ToUpper(s[0:1])+strings.ToLower(s[1:len(s)])
		col.Type=slt.mysqlToGo(DATA_TYPE)
		if EXTRA !=""{
		  if COLUMN_KEY=="PRI"{
		    col.Type=col.Type+` "orm:"pk;auto"`
			} else {col.Type=col.Type+`"orm:"auto"`
		  }
		}else{
		  if COLUMN_KEY=="PRI"{col.Type=col.Type+` "orm:"pk"`}		
		}
		//col.IsPrimary=COLUMN_KEY=="PRI"
		table.Columns=append(table.Columns,col)
        //log.Println(Column_Name,DATA_TYPE,COLUMN_KEY,EXTRA)
    }		
	slt.Tables=append(slt.Tables,table)  	
}

func GetCurrPath() string {
    file, _ := exec.LookPath(os.Args[0])
    path, _ := filepath.Abs(file)
    splitstring := strings.Split(path, "\\")
    size := len(splitstring)
    splitstring = strings.Split(path, splitstring[size-1])
    ret := strings.Replace(splitstring[0], "\\", "/", size-1)
    return ret
}

func main() {
	SqlStruct,err:=NewSqlStruct("my.go")
	defer SqlStruct.sqldb.Close()
	if err !=nil{
	  os.Exit(0)	
	}
	GoPath:=GetCurrPath()+"go/" 
    _,err=os.Stat(GoPath)
	if err!=nil { os.Mkdir(GoPath,os.ModePerm) }
	
    fout,err := os.Create(GoPath+SqlStruct.GoFile)
    defer fout.Close()
    if err != nil {
        fmt.Println(SqlStruct.GoFile,err)
        return
    }
	SqlStruct.BuildTableStruct()

	tmpl := template.New("my tmp")
	tmpl,_ = tmpl.Parse(
	`package {{ .GoFile}} 
import{ "time",
   "github.com/astaxie/beego/orm"}
{{range .Tables}} 
type {{.Table_name}} struct{        
  {{range .Columns}}
       {{.Name}}:{{.Type}}
  {{end}}  }	         
{{end}} `)
    //fmt.Println(len(SqlStruct.Tables)) 
    tmpl.Execute(fout,SqlStruct)
   	
}