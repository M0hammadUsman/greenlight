from locust import FastHttpUser, task


class MovieAPIUser(FastHttpUser):
    @task(1)  # Higher weight for this task
    def get_movies(self):
        self.client.get("/v1/movies?genres=adventure&page=1&page_size=2")

    @task(1)  # Lower weight for this task
    def healthcheck(self):
        self.client.get("/v1/healthcheck")
