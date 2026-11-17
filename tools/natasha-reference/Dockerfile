FROM python:3.15.2@sha256:46837ed3105db0029b6d656816185c6258e3b39690532fb782428e0a5fd76182

WORKDIR /work

COPY tools/natasha-reference/requirements.lock /tmp/requirements.lock

RUN python -m pip install --disable-pip-version-check --require-hashes -r /tmp/requirements.lock

ENTRYPOINT ["python", "tools/natasha-reference/reference.py"]
