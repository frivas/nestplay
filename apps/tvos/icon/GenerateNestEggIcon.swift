#!/usr/bin/env swift
import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

private enum Layer: String, CaseIterable {
  case back = "Back"
  case middle = "Middle"
  case front = "Front"
}

private struct Size {
  let width: Int
  let height: Int
  let suffix: String
}

private let canvas = CGSize(width: 400, height: 240)
private let smallSizes = [
  Size(width: 400, height: 240, suffix: ""),
  Size(width: 800, height: 480, suffix: "@2x"),
]
private let largeSizes = [Size(width: 1280, height: 768, suffix: "")]

private func srgb(_ hex: UInt32, alpha: CGFloat = 1) -> CGColor {
  CGColor(
    srgbRed: CGFloat((hex >> 16) & 0xFF) / 255,
    green: CGFloat((hex >> 8) & 0xFF) / 255,
    blue: CGFloat(hex & 0xFF) / 255,
    alpha: alpha)
}

private func quad(
  _ ctx: CGContext, from start: CGPoint, control: CGPoint, to end: CGPoint, color: UInt32,
  width: CGFloat
) {
  ctx.setStrokeColor(srgb(color))
  ctx.setLineWidth(width)
  ctx.setLineCap(.round)
  ctx.setLineJoin(.round)
  ctx.move(to: start)
  ctx.addQuadCurve(to: end, control: control)
  ctx.strokePath()
}

private func drawBack(_ ctx: CGContext) {
  let space = CGColorSpace(name: CGColorSpace.sRGB)!
  let gradient = CGGradient(
    colorsSpace: space,
    colors: [srgb(0x1B4148), srgb(0x0B1E22)] as CFArray,
    locations: [0, 1])!
  ctx.drawLinearGradient(
    gradient,
    start: CGPoint(x: 200, y: 0),
    end: CGPoint(x: 200, y: 240),
    options: [.drawsBeforeStartLocation, .drawsAfterEndLocation])
  ctx.setFillColor(srgb(0x081416))
  ctx.fillEllipse(in: CGRect(x: -10, y: 158, width: 420, height: 140))
  ctx.setFillColor(srgb(0x3D6A70))
  ctx.fillEllipse(in: CGRect(x: 43.8, y: 29.8, width: 4.4, height: 4.4))
  ctx.fillEllipse(in: CGRect(x: 76.4, y: 16.4, width: 3.2, height: 3.2))
  ctx.fillEllipse(in: CGRect(x: 350, y: 44, width: 4, height: 4))
  ctx.fillEllipse(in: CGRect(x: 326.6, y: 12.6, width: 2.8, height: 2.8))
  ctx.setFillColor(srgb(0x2F585E))
  ctx.fillEllipse(in: CGRect(x: 362.2, y: 106.2, width: 3.6, height: 3.6))
}

private func drawMiddle(_ ctx: CGContext) {
  ctx.setFillColor(srgb(0x4A3010))
  ctx.fillEllipse(in: CGRect(x: 82, y: 118, width: 236, height: 80))
  ctx.setFillColor(srgb(0x2B1A08))
  ctx.fillEllipse(in: CGRect(x: 108, y: 126, width: 184, height: 52))
  quad(
    ctx, from: CGPoint(x: 92, y: 136), control: CGPoint(x: 200, y: 192),
    to: CGPoint(x: 308, y: 138), color: 0x8A5A14, width: 18)
  quad(
    ctx, from: CGPoint(x: 84, y: 148), control: CGPoint(x: 198, y: 206),
    to: CGPoint(x: 316, y: 144), color: 0xC4922A, width: 13)
  quad(
    ctx, from: CGPoint(x: 102, y: 128), control: CGPoint(x: 206, y: 174),
    to: CGPoint(x: 298, y: 132), color: 0xE8C85A, width: 10)
  quad(
    ctx, from: CGPoint(x: 110, y: 158), control: CGPoint(x: 210, y: 208),
    to: CGPoint(x: 290, y: 156), color: 0xD4A84A, width: 9)
  quad(
    ctx, from: CGPoint(x: 118, y: 122), control: CGPoint(x: 170, y: 116),
    to: CGPoint(x: 228, y: 158), color: 0xB8862A, width: 8)
  quad(
    ctx, from: CGPoint(x: 172, y: 160), control: CGPoint(x: 252, y: 114),
    to: CGPoint(x: 308, y: 128), color: 0xF0D36A, width: 8)
  quad(
    ctx, from: CGPoint(x: 98, y: 164), control: CGPoint(x: 148, y: 116),
    to: CGPoint(x: 210, y: 124), color: 0xA07020, width: 7)
  quad(
    ctx, from: CGPoint(x: 188, y: 118), control: CGPoint(x: 260, y: 156),
    to: CGPoint(x: 304, y: 166), color: 0xC9A038, width: 7)
}

private func drawFront(_ ctx: CGContext) {
  let play = CGMutablePath()
  play.move(to: CGPoint(x: 181, y: 84))
  play.addLine(to: CGPoint(x: 181, y: 146))
  play.addLine(to: CGPoint(x: 239, y: 115))
  play.closeSubpath()
  ctx.setFillColor(srgb(0xF6EED8))
  ctx.setStrokeColor(srgb(0xF6EED8))
  ctx.setLineWidth(12)
  ctx.setLineJoin(.round)
  ctx.setLineCap(.round)
  ctx.addPath(play)
  ctx.drawPath(using: .fillStroke)
}

private func render(layer: Layer, width: Int, height: Int) -> CGImage {
  let space = CGColorSpace(name: CGColorSpace.sRGB)!
  let ctx = CGContext(
    data: nil,
    width: width,
    height: height,
    bitsPerComponent: 8,
    bytesPerRow: 0,
    space: space,
    bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
  ctx.translateBy(x: 0, y: CGFloat(height))
  ctx.scaleBy(x: CGFloat(width) / canvas.width, y: -CGFloat(height) / canvas.height)
  ctx.setShouldAntialias(true)
  ctx.interpolationQuality = .high
  switch layer {
  case .back: drawBack(ctx)
  case .middle: drawMiddle(ctx)
  case .front: drawFront(ctx)
  }
  return ctx.makeImage()!
}

private func writePNG(_ image: CGImage, to url: URL) {
  try? FileManager.default.createDirectory(
    at: url.deletingLastPathComponent(), withIntermediateDirectories: true)
  let dest = CGImageDestinationCreateWithURL(
    url as CFURL, UTType.png.identifier as CFString, 1, nil)!
  CGImageDestinationAddImage(dest, image, nil)
  precondition(CGImageDestinationFinalize(dest), "Failed to write \(url.path)")
}

private func imagesetURL(stack: String, layer: Layer) -> URL {
  brandRoot
    .appendingPathComponent("\(stack).imagestack")
    .appendingPathComponent("\(layer.rawValue).imagestacklayer")
    .appendingPathComponent("Content.imageset")
}

private let tvosRoot = URL(fileURLWithPath: #filePath)
  .deletingLastPathComponent()
  .deletingLastPathComponent()
private let brandRoot =
  tvosRoot
  .appendingPathComponent("Nestplay")
  .appendingPathComponent("Assets.xcassets")
  .appendingPathComponent("App Icon & Top Shelf Image.brandassets")

for layer in Layer.allCases {
  for size in smallSizes {
    let name = "\(layer.rawValue)\(size.suffix).png"
    writePNG(
      render(layer: layer, width: size.width, height: size.height),
      to: imagesetURL(stack: "App Icon - Small", layer: layer).appendingPathComponent(name))
  }
  writePNG(
    render(layer: layer, width: largeSizes[0].width, height: largeSizes[0].height),
    to: imagesetURL(stack: "App Icon - Large", layer: layer).appendingPathComponent(
      "\(layer.rawValue).png"))
}

print("Wrote Nest egg layers to \(brandRoot.path)")
