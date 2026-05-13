"""
GCS Emulator — Python SDK Compatibility Tests
Tests basic bucket/object CRUD against the running emulator.
"""
import sys
import os

from google.cloud import storage


EMULATOR_HOST = os.environ.get("GCS_EMULATOR_HOST", "http://localhost:9090")


def test_create_and_list_buckets():
    client = storage.Client(
        project="test-project",
        client_options={"api_endpoint": EMULATOR_HOST},
    )

    bucket_name = "python-sdk-bucket"
    bucket = client.create_bucket(bucket_name)
    assert bucket.name == bucket_name, f"Expected {bucket_name}, got {bucket.name}"

    buckets = list(client.list_buckets())
    assert any(b.name == bucket_name for b in buckets), "Bucket not found in list"


def test_upload_and_download_object():
    client = storage.Client(
        project="test-project",
        client_options={"api_endpoint": EMULATOR_HOST},
    )

    bucket_name = "python-io-bucket"
    bucket = client.bucket(bucket_name)
    if not bucket.exists():
        client.create_bucket(bucket_name)

    content = "hello from python sdk"
    blob = bucket.blob("test.txt")
    blob.upload_from_string(content)

    downloaded = blob.download_as_text()
    assert downloaded == content, f"Expected '{content}', got '{downloaded}'"


def test_list_objects_with_prefix():
    client = storage.Client(
        project="test-project",
        client_options={"api_endpoint": EMULATOR_HOST},
    )

    bucket_name = "python-list-bucket"
    bucket = client.bucket(bucket_name)
    if not bucket.exists():
        client.create_bucket(bucket_name)

    bucket.blob("logs/2024/jan.txt").upload_from_string("jan")
    bucket.blob("logs/2024/feb.txt").upload_from_string("feb")

    blobs = list(client.list_blobs(bucket_name, prefix="logs/2024/"))
    names = [b.name for b in blobs]
    assert "logs/2024/jan.txt" in names


if __name__ == "__main__":
    failures = []
    for name, fn in [
        ("create_and_list_buckets", test_create_and_list_buckets),
        ("upload_and_download_object", test_upload_and_download_object),
        ("list_objects_with_prefix", test_list_objects_with_prefix),
    ]:
        try:
            fn()
            print(f"  PASS: {name}")
        except Exception as e:
            print(f"  FAIL: {name} — {e}")
            failures.append(name)

    if failures:
        print(f"\n{failures} failed")
        sys.exit(1)
    else:
        print("\nAll Python SDK tests passed")
