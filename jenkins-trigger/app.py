import requests
from flask import Flask, render_template, request

app = Flask(__name__)

# ── Hardcoded Jenkins configuration ──────────────────────────────────
JENKINS_URL = "http://localhost:8081"
JENKINS_JOB = "webapp-go-deploy"
JENKINS_USER = "DiegoA"
JENKINS_TOKEN = "115fbdc6564aa974a64a0b93145b767db6"


def _build_url():
    return f"{JENKINS_URL}/job/{JENKINS_JOB}/buildWithParameters"


@app.route("/")
def index():
    return render_template("index.html", result=None)


@app.route("/trigger", methods=["POST"])
def trigger():
    client_name = request.form.get("client_name", "").strip()
    slug = request.form.get("slug", "").strip()

    if not client_name or not slug:
        result = {"ok": False, "msg": "Todos los campos son obligatorios."}
        return render_template("index.html", result=result,
                               client_name=client_name, slug=slug)

    params = {
        "CLIENT_NAME": client_name,
        "SLUG": slug,
    }

    try:
        resp = requests.post(
            _build_url(),
            auth=(JENKINS_USER, JENKINS_TOKEN),
            params=params,
            timeout=30,
        )
        if resp.status_code in (200, 201):
            result = {"ok": True, "msg": "Pipeline ejecutado exitosamente."}
        else:
            result = {
                "ok": False,
                "msg": f"Jenkins respondió con código {resp.status_code}.",
            }
    except requests.RequestException as exc:
        result = {"ok": False, "msg": f"No se pudo conectar a Jenkins: {exc}"}

    return render_template("index.html", result=result,
                           client_name=client_name, slug=slug)


if __name__ == "__main__":
    app.run(debug=False, port=5050)
