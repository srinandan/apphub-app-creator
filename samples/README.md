# apphub-app-creator command Samples

The following table contains some examples of apphub-app-creator.

| Operations | Command |
|---|---|
| generate | `apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --locations us-east1 --label-key $label_key`|
| generate | `apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --label-key $tag_key --tag-value $tag_value`|
| generate | `apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --per-k8s-namespace=true`|
| generate | `apphub-app-creator apps generate --parent folders/$folder --management-project $mp --locations us-west1 --log-label-key $log_label_key --log-label-value $log_label_value`|
| generate | `apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --per-k8s-app-label=true`|
| generate | `apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --label-key $label_key --report-only=true`|
| generate | `apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --auto-detect=true --report-only=true`|
| generate | `apphub-app-creator apps generate --parent folders/$folder --management-project $mp --locations us-west1 --project-keys proj1 --project-keys proj2 --app-name my-app`|
| delete   | `apphub-app-creator apps delete --management-project $project --locations us-west1 --locations us-east1`|
| delete   | `apphub-app-creator apps delete --name $name --management-project $project --locations us-west1`|


NOTE: This file is auto-generated during a release. Do not modify.