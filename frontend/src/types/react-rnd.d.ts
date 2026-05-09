declare module 'react-rnd' {
  import * as React from 'react'
  export interface Props {
    [key: string]: unknown
  }
  export class Rnd extends React.Component<Props> {}
}
