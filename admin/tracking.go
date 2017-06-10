package admin

import (
	"fmt"

	"bloodtales/system"
	"bloodtales/util"
	"bloodtales/models"
)

func handleAdminTracking() {
	handleAdminTemplate("/admin/trackings", system.TokenAuthentication, ViewTrackings, "trackings.tmpl.html")
	handleAdminTemplate("/admin/trackings/view", system.TokenAuthentication, ViewTracking, "tracking.tmpl.html")
	handleAdminTemplate("/admin/trackings/delete", system.TokenAuthentication, DeleteTracking, "")
}

func ViewTrackings(context *util.Context) {
	// paginate players query
	pagination, err := context.Paginate(context.DB.C(models.TrackingCollectionName).Find(nil).Sort("-t0"), DefaultPageSize)
	util.Must(err)

	// get resulting trackings
	var trackings []*models.Tracking
	util.Must(pagination.All(&trackings))

	// set template bindings
	context.Params.Set("trackings", trackings)
}

func ViewTracking(context *util.Context) {
	// parse parameters
	trackingId := context.Params.GetRequiredId("trackingId")

	tracking, err := models.GetTrackingById(context, trackingId)
	util.Must(err)
	
	// set template bindings
	context.Params.Set("tracking", tracking)
}

func DeleteTracking(context *util.Context) {
	// parse parameters
	trackingId := context.Params.GetRequiredId("trackingId")
	page := context.Params.GetInt("page", 1)

	tracking, err := models.GetTrackingById(context, trackingId)
	util.Must(err)

	tracking.Delete(context)

	context.Redirect(fmt.Sprintf("/admin/trackings?page=%d", page), 302)
}
