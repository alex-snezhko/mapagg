export type Response<T> = {
  success: true;
  data: T;
} | {
  success: false;
  error: string;
}

export type TagsResponse = Response<string[]>;
export type MapResponse = Response<[number, number, number][]>;

export interface AggregateDataTag {
  tag: string;
  weight: number;
}

export interface AggregateDataRequest {
  tags: AggregateDataTag[];
}

