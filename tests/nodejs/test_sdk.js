/**
 * GCS Emulator — Node.js SDK Compatibility Tests
 * Tests basic bucket/object CRUD against the running emulator.
 */
const { Storage } = require("@google-cloud/storage");

const EMULATOR_HOST = process.env.GCS_EMULATOR_HOST || "http://localhost:9090";

const storage = new Storage({
  apiEndpoint: EMULATOR_HOST,
  projectId: "test-project",
  credentials: {
    client_email: "test@test.iam.gserviceaccount.com",
    private_key: "test",
  },
});

async function testCreateAndListBuckets() {
  const bucketName = "nodejs-sdk-bucket";
  await storage.createBucket(bucketName);

  const [buckets] = await storage.getBuckets();
  const found = buckets.some((b) => b.name === bucketName);
  if (!found) throw new Error(`Bucket ${bucketName} not found`);
}

async function testUploadAndDownloadObject() {
  const bucketName = "nodejs-io-bucket";
  const bucket = storage.bucket(bucketName);
  const [exists] = await bucket.exists();
  if (!exists) await storage.createBucket(bucketName);

  const content = "hello from nodejs sdk";
  await bucket.file("test.txt").save(content);

  const [downloaded] = await bucket.file("test.txt").download();
  if (downloaded.toString() !== content) {
    throw new Error(`Expected '${content}', got '${downloaded}'`);
  }
}

async function testListObjectsWithPrefix() {
  const bucketName = "nodejs-list-bucket";
  const bucket = storage.bucket(bucketName);
  const [exists] = await bucket.exists();
  if (!exists) await storage.createBucket(bucketName);

  await bucket.file("logs/2024/jan.txt").save("jan");
  await bucket.file("logs/2024/feb.txt").save("feb");

  const [files] = await bucket.getFiles({ prefix: "logs/2024/" });
  const names = files.map((f) => f.name);
  if (!names.includes("logs/2024/jan.txt")) throw new Error("Object not found");
}

async function cleanup() {
  const [buckets] = await storage.getBuckets();
  for (const bucket of buckets) {
    const [files] = await bucket.getFiles();
    for (const file of files) {
      await file.delete();
    }
    await bucket.delete();
  }
}

(async () => {
  const tests = [
    ["create_and_list_buckets", testCreateAndListBuckets],
    ["upload_and_download_object", testUploadAndDownloadObject],
    ["list_objects_with_prefix", testListObjectsWithPrefix],
  ];

  const failures = [];
  for (const [name, fn] of tests) {
    try {
      await fn();
      console.log(`  PASS: ${name}`);
    } catch (e) {
      console.log(`  FAIL: ${name} — ${e.message}`);
      failures.push(name);
    }
  }

  try { await cleanup(); } catch {}

  if (failures.length) {
    console.log(`\n${failures} failed`);
    process.exit(1);
  } else {
    console.log("\nAll Node.js SDK tests passed");
  }
})();
