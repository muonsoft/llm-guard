FROM python:3.6.15@sha256:f8652afaf88c25f0d22354d547d892591067aa4026a7fa9a6819df9f300af6fc

WORKDIR /work

COPY tools/natasha-reference/requirements.lock /tmp/requirements.lock

RUN python -m pip install --disable-pip-version-check --require-hashes -r /tmp/requirements.lock

ENTRYPOINT ["python", "tools/natasha-reference/reference.py"]
