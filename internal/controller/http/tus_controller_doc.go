package http

// TUS Protocol Documentation
//
// This controller implements the TUS (Resumable Upload Protocol) v1.0.0
// Specification: https://tus.io/protocols/resumable-upload.html
//
// IMPORTANT: TUS protocol responses differ from standard REST API responses.
// This is required by the TUS protocol specification and cannot be changed.
//
// ========================================================================
// PROTOCOL DIFFERENCES: TUS vs REST
// ========================================================================
//
// REST API Responses:
// - Always return JSON with {success, message, code, data, timestamp}
// - Use helper.SendSuccessResponse(), helper.SendErrorResponse()
// - Status codes: 200, 201, 400, 401, 403, 404, 409, 500
//
// TUS Protocol Responses:
// - Upload initiation: JSON response with upload metadata (uses SendTusInitiateResponse)
// - Chunk upload: 204 No Content with headers (uses SendTusChunkResponse)
// - HEAD requests: 200 OK with headers only (uses SendTusHeadResponse)
// - DELETE requests: 204 No Content (uses SendTusDeleteResponse)
// - Errors: Protocol-specific status codes with headers (uses BuildTusErrorResponse)
//
// WHY THE DIFFERENCE?
// The TUS protocol specification requires specific response formats:
// - POST /upload (initiate): Returns 201 with Location header
// - PATCH /upload/{id} (chunk): Returns 204 with Upload-Offset header
// - HEAD /upload/{id}: Returns 200 with Upload-Offset and Upload-Length headers
// - DELETE /upload/{id}: Returns 204 No Content
//
// Standard JSON responses would break TUS client compatibility.
//
// ========================================================================
// TUS PROTOCOL FLOW
// ========================================================================
//
// 1. INITIATE UPLOAD (POST)
//    Client -> POST /api/v1/upload
//    Headers:
//      - Tus-Resumable: 1.0.0
//      - Upload-Length: [total file size in bytes]
//      - Upload-Metadata: [base64 encoded metadata]
//
//    Server -> 201 Created
//    Headers:
//      - Tus-Resumable: 1.0.0
//      - Location: [upload URL]
//      - Upload-Offset: 0
//    Body: JSON with upload_id, upload_url, offset, length
//
// 2. UPLOAD CHUNKS (PATCH)
//    Client -> PATCH [upload URL from Location header]
//    Headers:
//      - Tus-Resumable: 1.0.0
//      - Upload-Offset: [current offset]
//      - Content-Type: application/offset+octet-stream
//      - Content-Length: [chunk size]
//    Body: [binary chunk data]
//
//    Server -> 204 No Content (on success)
//    Headers:
//      - Tus-Resumable: 1.0.0
//      - Upload-Offset: [new offset after this chunk]
//
//    Server -> 409 Conflict (on offset mismatch)
//    Headers:
//      - Tus-Resumable: 1.0.0
//      - Upload-Offset: [expected offset]
//
// 3. CHECK UPLOAD STATUS (HEAD)
//    Client -> HEAD [upload URL]
//    Headers:
//      - Tus-Resumable: 1.0.0
//
//    Server -> 200 OK
//    Headers:
//      - Tus-Resumable: 1.0.0
//      - Upload-Offset: [current offset]
//      - Upload-Length: [total file size]
//
// 4. CANCEL UPLOAD (DELETE)
//    Client -> DELETE [upload URL]
//    Headers:
//      - Tus-Resumable: 1.0.0
//
//    Server -> 204 No Content
//
// ========================================================================
// ERROR HANDLING
// ========================================================================
//
// TUS protocol errors use specific HTTP status codes:
//
// 400 Bad Request
//   - Missing required headers (Tus-Resumable, Upload-Length, etc.)
//   - Invalid header values
//   - Malformed metadata
//
// 401 Unauthorized
//   - Missing or invalid authentication
//
// 403 Forbidden
//   - User doesn't have access to the upload
//
// 404 Not Found
//   - Upload ID doesn't exist
//
// 409 Conflict
//   - Upload offset mismatch (client sent wrong offset)
//   - Upload already completed
//   - Response includes Upload-Offset header with expected value
//
// 412 Precondition Failed
//   - Unsupported TUS protocol version
//
// 413 Request Entity Too Large
//   - File size exceeds maximum
//   - Chunk size exceeds maximum
//
// 423 Locked
//   - Upload session is inactive/expired
//
// 500 Internal Server Error
//   - Server-side processing error
//
// ALL error responses include Tus-Resumable header for protocol compliance.
//
// ========================================================================
// HELPER ENDPOINTS (Non-TUS)
// ========================================================================
//
// These endpoints follow standard REST API patterns:
//
// - GET /api/v1/upload/slot/check - Check upload slot availability
// - POST /api/v1/upload/queue/reset - Reset upload queue
// - GET /api/v1/upload/{id}/info - Get upload information
//
// These use standard JSON responses with {success, message, code, data, timestamp}.
//
// ========================================================================
// INTEGRATION EXAMPLES
// ========================================================================
//
// Example 1: Simple File Upload (curl)
// ------------------------------------
// # 1. Initiate upload
// curl -X POST http://localhost:3000/api/v1/upload \
//   -H "Tus-Resumable: 1.0.0" \
//   -H "Upload-Length: 1048576" \
//   -H "Upload-Metadata: filename dGVzdC5qcGc,is_video dHJ1ZQ==" \
//   -H "Authorization: Bearer {access_token}"
//
// # Response:
// {
//   "status": "success",
//   "message": "Upload berhasil diinisiasi",
//   "code": 201,
//   "data": {
//     "upload_id": "abc123",
//     "upload_url": "http://localhost:3000/api/v1/upload/abc123",
//     "offset": 0,
//     "length": 1048576
//   },
//   "timestamp": "2024-01-01T10:00:00Z"
// }
//
// # 2. Upload chunk
// curl -X PATCH http://localhost:3000/api/v1/upload/abc123 \
//   -H "Tus-Resumable: 1.0.0" \
//   -H "Upload-Offset: 0" \
//   -H "Content-Type: application/offset+octet-stream" \
//   -H "Content-Length: 524288" \
//   -H "Authorization: Bearer {access_token}" \
//   --data-binary @chunk1.bin
//
// # Response: 204 No Content
// # Headers: Tus-Resumable: 1.0.0, Upload-Offset: 524288
//
// # 3. Upload remaining chunk
// curl -X PATCH http://localhost:3000/api/v1/upload/abc123 \
//   -H "Tus-Resumable: 1.0.0" \
//   -H "Upload-Offset: 524288" \
//   -H "Content-Type: application/offset+octet-stream" \
//   -H "Content-Length: 524288" \
//   -H "Authorization: Bearer {access_token}" \
//   --data-binary @chunk2.bin
//
// # Response: 204 No Content
// # Headers: Tus-Resumable: 1.0.0, Upload-Offset: 1048576
//
// Example 2: Resumable Upload (tus-js-client)
// -------------------------------------------
// import { Upload } from 'tus-js-client';
//
// const file = document.querySelector('#file-input').files[0];
//
// const upload = new Upload(file, {
//   endpoint: 'http://localhost:3000/api/v1/upload',
//   chunkSize: 5 * 1024 * 1024, // 5MB chunks
//   metadata: {
//     filename: file.name,
//     filetype: file.type
//   },
//   headers: {
//     'Authorization': `Bearer ${accessToken}`
//   },
//   onError: (error) => {
//     console.error('Upload failed:', error);
//   },
//   onProgress: (bytesSent, bytesTotal) => {
//     const percentage = (bytesSent / bytesTotal * 100).toFixed(2);
//     console.log(`Upload progress: ${percentage}%`);
//   },
//   onSuccess: () => {
//     console.log('Upload finished!');
//   }
// });
//
// upload.start();
//
// Example 3: Check Upload Status
// -------------------------------
// curl -I http://localhost:3000/api/v1/upload/abc123 \
//   -H "Tus-Resumable: 1.0.0" \
//   -H "Authorization: Bearer {access_token}"
//
// # Response: 200 OK
// # Headers:
// #   Tus-Resumable: 1.0.0
// #   Upload-Offset: 524288
// #   Upload-Length: 1048576
//
// Example 4: Cancel Upload
// ------------------------
// curl -X DELETE http://localhost:3000/api/v1/upload/abc123 \
//   -H "Tus-Resumable: 1.0.0" \
//   -H "Authorization: Bearer {access_token}"
//
// # Response: 204 No Content
//
// ========================================================================
// PROJECT UPDATE UPLOADS
// ========================================================================
//
// Project update uploads follow the same TUS protocol but are scoped to a project:
//
// - POST /api/v1/project/{id}/upload - Initiate project update
// - PATCH /api/v1/project/{id}/upload/{upload_id} - Upload chunk
// - HEAD /api/v1/project/{id}/upload/{upload_id} - Check status
// - GET /api/v1/project/{id}/upload/{upload_id}/info - Get info (REST)
// - DELETE /api/v1/project/{id}/upload/{upload_id} - Cancel upload
//
// The TUS protocol behavior is identical, but uploads are scoped to a project.
//
// ========================================================================
// CONFIGURATION
// ========================================================================
//
// TUS protocol configuration (config.Upload):
// - TusVersion: "1.0.0" - Protocol version
// - ChunkSize: 1048576 (1MB) - Default chunk size
// - MaxConcurrent: 3 - Maximum concurrent uploads per user
// - MaxProjectFileSize: 524288000 (500MB) - Max file size for projects
// - MaxModulFileSize: 52428800 (50MB) - Max file size for modules
//
// ========================================================================
// TESTING
// ========================================================================
//
// When testing TUS endpoints:
// 1. Always include Tus-Resumable header
// 2. Follow the protocol flow (initiate -> patch -> head -> delete)
// 3. Check response headers (not just status code)
// 4. Handle offset mismatches gracefully (client should retry with correct offset)
// 5. Test resumability (pause upload, resume with HEAD, continue with PATCH)
//
// ========================================================================
// REFERENCES
// ========================================================================
//
// - TUS Protocol Specification: https://tus.io/protocols/resumable-upload.html
// - TUS Client Libraries: https://tus.io/implementations.html
// - tUS-JS-Client: https://github.com/tus/tus-js-client
// - PyTUS (Python): https://github.com/tus/py tus
//
// ========================================================================
