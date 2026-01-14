package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": "@request.auth.id = chat.user",
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_3861817060",
					"hidden": false,
					"id": "relation1704850090",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "chat",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text4274335913",
					"max": 0,
					"min": 0,
					"name": "content",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "select1466534506",
					"maxSelect": 1,
					"name": "role",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "select",
					"values": [
						"user",
						"ai"
					]
				},
				{
					"hidden": false,
					"id": "json3622966325",
					"maxSize": 0,
					"name": "meta",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": false,
					"type": "autodate"
				},
				{
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": false,
					"type": "autodate"
				}
			],
			"id": "pbc_2605467279",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_x6ZH7e7EAZ` + "`" + ` ON ` + "`" + `messages` + "`" + ` (` + "`" + `chat` + "`" + `)"
			],
			"listRule": "@request.auth.id = chat.user",
			"name": "messages",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": "@request.auth.id = chat.user"
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_2605467279")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
