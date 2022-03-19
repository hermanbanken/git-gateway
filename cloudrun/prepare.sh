# enable firestore
gcloud services enable firestore.googleapis.com
# grant default Compute Service Account access to firestore
gcloud projects add-iam-policy-binding $(gcloud config get-value project) \
--member=serviceAccount:$(gcloud projects list --filter="$(gcloud config get-value project)" --format="value(PROJECT_NUMBER)")-compute@developer.gserviceaccount.com \
--role=roles/datastore.user
