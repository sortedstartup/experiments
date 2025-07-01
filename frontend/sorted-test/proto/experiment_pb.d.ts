// package: sortedtest
// file: experiment.proto

import * as jspb from "google-protobuf";

export class testRequest extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  getChatId(): string;
  setChatId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): testRequest.AsObject;
  static toObject(includeInstance: boolean, msg: testRequest): testRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: testRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): testRequest;
  static deserializeBinaryFromReader(message: testRequest, reader: jspb.BinaryReader): testRequest;
}

export namespace testRequest {
  export type AsObject = {
    message: string,
    chatId: string,
  }
}

export class testResponse extends jspb.Message {
  getText(): string;
  setText(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): testResponse.AsObject;
  static toObject(includeInstance: boolean, msg: testResponse): testResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: testResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): testResponse;
  static deserializeBinaryFromReader(message: testResponse, reader: jspb.BinaryReader): testResponse;
}

export namespace testResponse {
  export type AsObject = {
    text: string,
  }
}

