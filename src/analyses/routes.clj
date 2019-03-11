(ns analyses.routes
  (:use [common-swagger-api.schema.badges]
        [common-swagger-api.schema :only [StandardUserQueryParams]])
  (:require [compojure.api.sweet :refer :all]
            [common-swagger-api.schema.apps :refer [AnalysisSubmission]]
            [clojure-commons.exception :refer [exception-handlers]]
            [clojure-commons.lcase-params :refer [wrap-lcase-params]]
            [clojure-commons.query-params :refer [wrap-query-params]]
            [compojure.api.middleware :refer [wrap-exceptions]]
            [ring.util.http-response :refer [ok]]
            [ring.middleware.keyword-params :refer [wrap-keyword-params]]
            [ring.swagger.coerce :as rc]
            [ring.swagger.common :refer [value-of]]
            [schema.core :as s]
            [schema.coerce :as sc]
            [schema.utils :as su]
            [service-logging.middleware :refer [log-validation-errors add-user-to-context]]
            [slingshot.slingshot :refer [throw+]]
            [analyses.persistence :refer [add-badge get-badge update-badge delete-badge]])
  (:import [java.util UUID]))

(s/defschema DeletionResponse
  {:id (describe UUID "The UUID of the resource that was deleted")})

(defn coerce-string->long
  "When the given map contains the given key, converts its string value to a long."
  [m k]
  (if (contains? m k)
    (update m k rc/string->long)
    m))

(defn- stringify-uuids
  [v]
  (if (instance? UUID v)
    (str v)
    v))

(def ^:private custom-coercions {String stringify-uuids})

(defn- custom-coercion-matcher
  [schema]
  (or (rc/json-schema-coercion-matcher schema)
      (custom-coercions schema)))

(defn coerce
  [schema value]
  ((sc/coercer (value-of schema) custom-coercion-matcher) value))

(defn coerce!
  [schema value]
  (let [result (coerce schema value)]
    (if (su/error? result)
      (throw+ (assoc result :type :compojure.api.exception/response-validation))
      result)))

(defapi app
  (swagger-routes
   {:ui "/docs"
    :spec "/swagger.json"
    :data {:info {:title "Analyses API"
                  :description "Swaggerized Analyses API"}
           :tags [{:name "badges" :description "The API for managing badges."}]
           :consumes ["application/json"]
           :produces ["application/json"]}})

  (middleware
    [add-user-to-context
     wrap-query-params
     wrap-lcase-params
     wrap-keyword-params
     [wrap-exceptions exception-handlers]
     log-validation-errors]

    (GET "/" [] (ok (str "yo what up\n")))

    (context "/badges" []
      :tags ["badges"]

      (POST "/" []
        :body         [badge NewBadge]
        :query        [{:keys [user]} StandardUserQueryParams]
        :return       Badge
        :summary      "Adds a badge to the database"
        :description  "Adds a badge and corresponding submission information to the
        database. The username passed in should already exist. A new UUID will be
        assigned and returned."
        (ok (coerce! Badge (add-badge user badge))))

      (GET "/:id" [id]
        :return       Badge
        :query        [{:keys [user]} StandardUserQueryParams]
        :summary      "Gets badge information from the database"
        :description  "Gets the badge information from the database, including its
        UUID, the name of the user that owns it, and the submission JSON"
        (ok (coerce! Badge (get-badge id user))))

      (PATCH "/:id" [id]
        :body         [badge UpdateBadge]
        :query        [{:keys [user]} StandardUserQueryParams]
        :return       Badge
        :summary      "Modifies an existing badge"
        :description  "Modifies an existing badge, allowing the caller to change
        owners and the contents of the submission JSON"
        (ok (coerce! Badge (update-badge id user badge))))

      (DELETE "/:id" [id]
        :query        [{:keys [user]} StandardUserQueryParams]
        :return        DeletionResponse
        :summary      "Deletes a badge"
        :description  "Deletes a badge from the database. Will returns a success
        even if called on a badge that has either already been deleted or never
        existed in the first place"
        (ok (coerce! DeletionResponse (delete-badge id user)))))))
