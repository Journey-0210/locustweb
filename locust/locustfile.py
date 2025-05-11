from locust import HttpUser, task, between, events
import json
import os
import time

class WebsiteUser(HttpUser):
    wait_time = between(1, 2.5)

    @task
    def index(self):
        self.client.get("/")

# 收集自定义指标
class MetricsCollector:
    def __init__(self):
        self.total_requests = 0
        self.success_requests = 0
        self.failure_requests = 0
        self.total_response_time = 0
        self.max_response_time = 0
        self.min_response_time = float('inf')
        self.total_content_size = 0
        self.start_time = None
        self.end_time = None

    def on_request(self, request_type, name, response_time, response_length, response, context, exception, **kwargs):
        if self.start_time is None:
            self.start_time = time.time()

        self.total_requests += 1
        self.total_response_time += response_time
        self.max_response_time = max(self.max_response_time, response_time)
        self.min_response_time = min(self.min_response_time, response_time)

        if exception:
            self.failure_requests += 1
        else:
            self.success_requests += 1
            self.total_content_size += response_length

    def stop(self):
        self.end_time = time.time()
        duration = self.end_time - self.start_time if self.start_time else 1
        error_rate = self.failure_requests / self.total_requests if self.total_requests else 0
        tps = self.success_requests / duration if duration > 0 else 0
        download_speed = self.total_content_size / duration if duration > 0 else 0

        metrics = {
            "total_requests": self.total_requests,
            "success_requests": self.success_requests,
            "failure_requests": self.failure_requests,
            "avg_response_time": self.total_response_time / self.total_requests if self.total_requests else 0,
            "max_response_time": self.max_response_time,
            "min_response_time": self.min_response_time if self.min_response_time != float('inf') else 0,
            "error_rate": error_rate,
            "tps": tps,
            "download_speed": download_speed,
            "total_download_size": self.total_content_size,
            "total_duration": duration,
            "dns_time": 0,  # 如有需要请集成 browser instrumentation
            "connect_time": 0,
            "first_byte_time": 0,
            "content_time": 0,
            "availability": 1 - error_rate
        }

        output_dir = "results"
        os.makedirs(output_dir, exist_ok=True)
        with open(os.path.join(output_dir, "test_{}_metrics.json".format(int(time.time()))), "w") as f:
            json.dump(metrics, f, indent=2)

collector = MetricsCollector()

@events.request.add_listener
def on_request(**kwargs):
    collector.on_request(**kwargs)

@events.quitting.add_listener
def on_quit(environment, **kwargs):
    collector.stop()
