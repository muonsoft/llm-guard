FROM python:3.15.2@sha256:e31e8561e10880b4e2ce95a7b3d14c3bf006bc0a7934adccd407905290e77be0

WORKDIR /work

COPY tools/natasha-reference/requirements.lock /tmp/requirements.lock

RUN python -m pip install --disable-pip-version-check --require-hashes -r /tmp/requirements.lock

ENTRYPOINT ["python", "tools/natasha-reference/reference.py"]
