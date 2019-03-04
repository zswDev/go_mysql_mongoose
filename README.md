"灵感来源于 mysql_mongoose" 


example:

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
		"state":  1,
		"$eq":    S{"state", 1},
		"$sort":  S{"id", -1},
		"$limit": 1,
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
---------------------------
>>
select id,rank,state,name from tests where  (  (  (  ( id = ? )  and  ( state = ? )  )  or  (  ( id = ? )  and  ( state = ? )  )  )  and  ( state = ? ) and  ( state = ? ) )  order by id desc  limit 1 [1 1 2 1 1 1] 
[map[name:22222222a id:2 rank:2 state:1]]

insert into tests ( rank,state,name ) values ( ?,?,? )  [1 3 aab] 
[map[lastInserId:30 rowsAffected:1]]

delete from tests where  (  ( id = ? )  )  [14] 
[map[lastInserId:0 rowsAffected:0]]

update tests set name=? where  (  ( id = ? )  )  [aabc 15] 
[map[lastInserId:0 rowsAffected:0]]

 
			update tests
			set name="xxxx"
			where id=?
		 [1] 
[map[lastInserId:0 rowsAffected:0]]
```