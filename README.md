"灵感来源于 mysql_mongoose" 

```golang	var db DB
	db.Connetion(URL)
	defer db.Close()

	rows := db.Find("tests", M{
		"$or": S{
			M{
				"id":    1,
				"state": 1,
			},
			M{
				"id":    2,
				"state": 1,
			},
		},
	}, "id rank state name")
	fmt.Println(rows)

	rows = db.Insert("tests", M{
		"rank":  1,
		"state": 3,
		"name":  "aab",
	})
	fmt.Println(rows)

	rows = db.Remove("tests", M{
		"id": 14,
	})
	fmt.Println(rows)

	rows = db.Update("tests", M{
		"id": 15,
	}, M{
		"name": "aabc",
	})
	fmt.Println(rows)

	rows = db.Query(`
		update tests
		set name="xxxx"
		where id=?
	`, 1)
	fmt.Println(rows)

```