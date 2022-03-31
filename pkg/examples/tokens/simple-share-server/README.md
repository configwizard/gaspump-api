# NeoFS RESTful API

This API offers access to containers controlled by you or others, using the HTTP protocol. Behind the scenes, this uses gRPC to communicate directly with NeoFS

* You can use our testnet node here 
* You can use our mainnet node here


Quick start

1. Create a container manually. You can use the example script provided here.
This will create a container with `EACLBasicPublic` permissions. This means that its default open for read/write to anyone with a valid bearer token
2. Now you will want to add an EACL table to the container to restrict access so that no one (except you) can write to it
To do so, you create an EACL table and pin it to the container



```yaml
openapi: 3.0.0
info:
    version: 0.0.1
    title: greenfinch-api
    license:
        name: Apache 2.0
        url: https://www.apache.org/licenses/LICENSE-2.0.html
    description: An HTTP RESTful API
    termsOfService: TBA
    contact:
        name: Alex Walker
        url: discord.com
        email: amlwwalker@gmail.com
servers:
    -
        url: 'localhost:3000/api/v1'
paths:
    /bearer:
        get:
            summary: Returns an unsigned base64 encoded binary bearer token
            operationId: retrieveUnsignedBearerToken
            tags:
                - token
            responses:
                '201':
                    description: Created
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Bearer'
                '400':
                    description: Bad Request
                '502':
                    description: Server error
                default:
                    description: Bad Request Error
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Error'
    '/object/{containerId}/{objectId}':
        head:
            summary: Returns the meta data for a specified object
            operationId: retrieveObjectMetaData
            parameters:
                -
                    name: containerId
                    in: path
                    required: true
                    description: Container ID that the object resides under
                    schema:
                        type: string
                -
                    name: objectId
                    in: path
                    required: true
                    description: Object ID to retrieve metadata for
                    schema:
                        type: string
            tags:
                - object
            responses:
                '200':
                    description: OK
                    headers:
                        application/octet:
                            schema:
                                $ref: '#/components/schemas/Metadata'
                        NEOFS-META:
                            schema:
                                type: string
                '400':
                    description: Bad Request
                '502':
                    description: Server error
                default:
                    description: Bad Request Error
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Error'
        get:
            summary: Returns the binary content of an object
            operationId: retrieveObjectData
            parameters:
                -
                    name: containerId
                    in: path
                    required: true
                    description: Container ID that the object resides under
                    schema:
                        type: string
                -
                    name: objectId
                    in: path
                    required: true
                    description: Object ID to retrieve metadata for
                    schema:
                        type: string
            tags:
                - object
            responses:
                '200':
                    description: OK
                    headers:
                        NEOFS-META:
                            schema:
                                type: string
                    content:
                        application/octet:
                            schema:
                                $ref: '#/components/schemas/BinaryData'
                '400':
                    description: Bad Request
                '502':
                    description: Server error
                default:
                    description: Bad Request Error
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Error'
        delete:
            summary: Deletes the object from the container permanently
            operationId: deteleObject
            parameters:
                -
                    name: containerId
                    in: path
                    required: true
                    description: Container ID that the object resides under
                    schema:
                        type: string
                -
                    name: objectId
                    in: path
                    required: true
                    description: Object ID to retrieve metadata for
                    schema:
                        type: string
            tags:
                - object
            responses:
                '204':
                    description: OK
                '400':
                    description: Bad Request
                '502':
                    description: Server error
                default:
                    description: Bad Request Error
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Error'
    '/object/{containerId}':
        post:
            summary: Upload an object
            operationId: object
            parameters:
                -
                    name: containerId
                    in: path
                    required: true
                    description: Container ID that the object should reside under
                    schema:
                        type: string
            tags:
                - object
            responses:
                '201':
                    description: Upload successful
                    headers:
                        application/octet:
                            schema:
                                $ref: '#/components/schemas/Metadata'
                        NEOFS-META:
                            schema:
                                type: string
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Object'
                '400':
                    description: Bad Request
                '502':
                    description: Server error
                default:
                    description: Bad Request Error
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/Error'
components:
    schemas:
        Metadata:
            type: string
            format: string
        BinaryData:
            type: string
            format: binary
        Object:
            type: object
            required:
                - id
        Bearer:
            type: object
            required:
                - created_at
                - token
            properties:
                created_at:
                    type: object
                token:
                    type: string
        Error:
            type: object
            required:
                - code
                - message
            properties:
                code:
                    type: integer
                    format: int32
                message:
                    type: string

```
