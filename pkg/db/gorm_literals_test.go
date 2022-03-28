package db

var categoriesTestSeed = []Category{
	{Name: "Groceries"},
	{Name: "Food"},
	{Name: "Clothing"},
	{Name: "Snacking"},
}

var entriesTestSeed []Entry

// requiring db running, thus cannot be init()
func dynInitEntriesTestSeed() {

	var pc *Category // works for pointer method ByName

	//
	entriesTestSeed = []Entry{

		// id 1-3
		{
			Name:     "Soap",
			Comment:  "new cat - name exists",
			Category: Category{Name: "Groceries"}, // fails
		},
		{
			Name:    "Toothpaste",
			Comment: "new cat - name not exists",
			// Category: Category{Name: fmt.Sprintf("Groceries-%v", time.Now().Unix())},
			Category: Category{Name: "Groceries-2"},
		},
		{
			Name:       "WC Cleaner",
			Comment:    "cat by ID",
			CategoryID: pc.ByName("Groceries"),
		},

		// id 4,5
		{
			Name:       "Coffee",
			Comment:    "two new credit cards",
			CategoryID: pc.ByName("Snacking"),
			CreditCards: []CreditCard{
				{Issuer: "VISA", Number: 232233339090},
				{Issuer: "AMEX", Number: 909090909090},
			},
		},
		{
			Name:       "Cookie",
			Comment:    "same new credit card - independent",
			CategoryID: pc.ByName("Snacking"),
			CreditCards: []CreditCard{
				{Issuer: "VISA", Number: 232233339090}, // gets duplicated
			},
		},

		// explicit ID
		{
			ID:         uint(13),
			Name:       "Apple Pie",
			CategoryID: pc.ByName("Snacking"),
		},
		{
			ID:         uint(14),
			Name:       "Nougat",
			Comment:    "three new tags",
			CategoryID: pc.ByName("Snacking"),
			Tags: []Tag{
				{Name: "Indulgence"},
				{Name: "Reward"},
				{Name: "Craving"},
			},
		},
		{
			ID:         uint(15),
			Name:       "Marzipan",
			Comment:    "three new tags again",
			CategoryID: pc.ByName("Snacking"),
			Tags: []Tag{
				{Name: "Indulgence"}, // neither inserted (unique) nor omitted; m-n table contains wrong entry with TagID 0
				{Name: "Reward"},
				{Name: "Craving"},
				{Name: "Sloth"}, // is this added, though the others fail?
			},
		},

		// xxx
	}

}
