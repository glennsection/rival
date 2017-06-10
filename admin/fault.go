package admin

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"bloodtales/system"
	"bloodtales/models"
	"bloodtales/util"
)

func handleAdminFaults() {
	handleAdminTemplate("/admin/faults", system.TokenAuthentication, ViewFaults, "faults.tmpl.html")
	handleAdminTemplate("/admin/faults/view", system.TokenAuthentication, ViewFault, "fault.tmpl.html")
	handleAdminTemplate("/admin/faults/delete", system.TokenAuthentication, DeleteFault, "")
}

func ViewFaults(context *util.Context) {
	// parse parameters
	search := context.Params.GetString("search", "")

	// process search terms
	var query *mgo.Query = nil
	if search != "" {
		// find users matching name search
		var users []*models.User
		util.Must(context.DB.C(models.UserCollectionName).Find(bson.M {
			"nm": bson.M {
				"$regex": bson.RegEx {
					Pattern: fmt.Sprintf(".*%s.*", search),
					Options: "i",
				},
			},
		}).All(&users))
		userIDs := make([]bson.ObjectId, len(users))
		for i, user := range users {
			userIDs[i] = user.ID
		}

		// build query
		query = context.DB.C(util.FaultCollectionName).Find(bson.M {
			"$or": []bson.M {
				bson.M { "us":
					bson.M {
						"$in": userIDs,
					},
				},
				bson.M { "err":
					bson.M {
						"$regex": bson.RegEx {
							Pattern: fmt.Sprintf(".*%s.*", search),
							Options: "i",
						},
					},
				},
			},
		})
	} else {
		query = context.DB.C(util.FaultCollectionName).Find(nil)
	}

	// sorting
	query = context.Sort(query, "t0-desc")

	// paginate users query
	pagination, err := context.Paginate(query, DefaultPageSize)
	util.Must(err)

	// get resulting faults
	var faults []*util.Fault
	util.Must(pagination.All(&faults))

	// set template bindings
	context.Params.Set("faults", faults)
}

func ViewFault(context *util.Context) {
	// parse parameters
	faultId := context.Params.GetRequiredId("faultId")

	fault, err := util.GetFaultById(context, faultId)
	util.Must(err)
	
	// set template bindings
	context.Params.Set("fault", fault)
}

func DeleteFault(context *util.Context) {
	// parse parameters
	faultId := context.Params.GetRequiredId("faultId")
	page := context.Params.GetInt("page", 1)

	fault, err := util.GetFaultById(context, faultId)
	util.Must(err)

	fault.Delete(context)

	context.Redirect(fmt.Sprintf("/admin/faults?page=%d", page), 302)
}
