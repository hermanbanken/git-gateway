if $GOOGLE_CLOUD_REGION == 'europe-west4'; then
    echo "Pick a region that hosts App Engine"
    exit 1
fi

export GOOGLE_CLOUD_PROJECT_ID=$(gcloud projects list --filter="$(gcloud config get-value project)" --format="value(PROJECT_NUMBER)")

# setup secret
gcloud services enable secretmanager.googleapis.com --project=$GOOGLE_CLOUD_PROJECT
export SECRET=git-gateway
gcloud secrets create $SECRET --replication-policy="automatic"
gcloud secrets add-iam-policy-binding $SECRET \
--member=serviceAccount:$GOOGLE_CLOUD_PROJECT_ID-compute@developer.gserviceaccount.com \
--role=roles/secretmanager.secretAccessor \
--project $GOOGLE_CLOUD_PROJECT

# prepare bootstrap
gcloud secrets add-iam-policy-binding $SECRET \
--member=serviceAccount:$GOOGLE_CLOUD_PROJECT_ID-compute@developer.gserviceaccount.com \
--role=roles/secretmanager.secretVersionAdder \
--project $GOOGLE_CLOUD_PROJECT

# cleanup after initialization
gcloud secrets remove-iam-policy-binding $SECRET \
--member=serviceAccount:$GOOGLE_CLOUD_PROJECT_ID-compute@developer.gserviceaccount.com \
--role=roles/secretmanager.secretVersionAdder \
--project $GOOGLE_CLOUD_PROJECT

# enable firestore (TODO $GOOGLE_CLOUD_REGION does not work here => europe-west4 has no App Engine => must be europe-west1 belgium/europe-west2 london)
gcloud services enable firestore.googleapis.com appengine.googleapis.com --project=$GOOGLE_CLOUD_PROJECT
gcloud app create --region $GOOGLE_CLOUD_REGION --project=$GOOGLE_CLOUD_PROJECT
gcloud alpha firestore databases create --region $GOOGLE_CLOUD_REGION --project=$GOOGLE_CLOUD_PROJECT

# grant default Compute Service Account access to firestore
gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT \
--member=serviceAccount:$GOOGLE_CLOUD_PROJECT_ID-compute@developer.gserviceaccount.com \
--role=roles/datastore.user

# enable bootstrap mode
## TODO replace with secret check
gcloud run services update $K_SERVICE --region $GOOGLE_CLOUD_REGION --project=$GOOGLE_CLOUD_PROJECT --platform managed --args=bootstrap
