FROM python:3.15.2@sha256:e4bc83a3314d078d651f06c8440e9a03ba892ee2f6577437f174abd8731fb343

WORKDIR /work

COPY tools/natasha-reference/requirements.lock /tmp/requirements.lock

RUN python -m pip install --disable-pip-version-check --require-hashes -r /tmp/requirements.lock

ENTRYPOINT ["python", "tools/natasha-reference/reference.py"]
