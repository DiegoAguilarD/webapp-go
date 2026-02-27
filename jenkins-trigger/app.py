import requests
from flask import Flask, render_template, redirect, url_for

app = Flask(__name__)

# ── Hardcoded Jenkins configuration ──────────────────────────────────
JENKINS_URL = "http://localhost:8080"
JENKINS_JOB = "webapp-go"
JENKINS_USER = "admin"
JENKINS_TOKEN = "CHANGE_ME"

# ── Hardcoded pipeline parameters ────────────────────────────────────
PIPELINE_PARAMS = {
    "CLIENT_NAME": "Acme Corp",
    "SLUG": "acme_corp",
    "ADMIN_EMAIL": "admin@acme.com",
    "ADMIN_PASSWORD": "changeme123",
}


def _build_url():
    return f"{JENKINS_URL}/job/{JENKINS_JOB}/buildWithParameters"


@app.route("/")
def index():
    return render_template("index.html", params=PIPELINE_PARAMS, result=None)


@app.route("/trigger", methods=["POST"])
def trigger():
    try:
        resp = requests.post(
            _build_url(),
            auth=(JENKINS_USER, JENKINS_TOKEN),
            params=PIPELINE_PARAMS,
            timeout=30,
        )
        if resp.status_code in (200, 201):
            result = {"ok": True, "msg": "Pipeline ejecutado exitosamente."}
        else:
            result = {
                "ok": False,
                "msg": f"Error {resp.status_code}: {resp.text[:200]}",
            }
    except requests.RequestException as exc:
        result = {"ok": False, "msg": f"No se pudo conectar a Jenkins: {exc}"}

    return render_template("index.html", params=PIPELINE_PARAMS, result=result)


if __name__ == "__main__":
    app.run(debug=True, port=5050)
