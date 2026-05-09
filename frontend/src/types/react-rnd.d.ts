declare module 'react-rnd' {
  import { Component } from 'react'

  interface Position {
    x: number
    y: number
  }

  interface Size {
    width: number | string
    height: number | string
  }

  interface DraggableData {
    node: HTMLElement
    x: number
    y: number
    deltaX: number
    deltaY: number
    lastX: number
    lastY: number
  }

  interface RndProps {
    style?: React.CSSProperties
    className?: string
    bounds?: string
    position?: Position
    size?: Size
    default?: { x?: number; y?: number; width?: number | string; height?: number | string }
    minWidth?: number | string
    minHeight?: number | string
    maxWidth?: number | string
    maxHeight?: number | string
    z?: number
    dragGrid?: [number, number]
    resizeGrid?: [number, number]
    lockAspectRatio?: boolean
    enableResizing?: boolean | { top?: boolean; right?: boolean; bottom?: boolean; left?: boolean; topRight?: boolean; bottomRight?: boolean; bottomLeft?: boolean; topLeft?: boolean }
    disableDragging?: boolean
    onDragStart?: (e: MouseEvent | TouchEvent, data: DraggableData) => void
    onDrag?: (e: MouseEvent | TouchEvent, data: DraggableData) => void
    onDragStop?: (e: MouseEvent | TouchEvent, data: DraggableData) => void
    onResizeStart?: (e: MouseEvent | TouchEvent, dir: string, ref: HTMLElement) => void
    onResize?: (e: MouseEvent | TouchEvent, dir: string, ref: HTMLElement, delta: { width: number; height: number }, position: Position) => void
    onResizeStop?: (e: MouseEvent | TouchEvent, dir: string, ref: HTMLElement, delta: { width: number; height: number }, position: Position) => void
    onClick?: (e: React.MouseEvent) => void
    children?: React.ReactNode
  }

  export class Rnd extends Component<RndProps> {}
  export default Rnd
}
