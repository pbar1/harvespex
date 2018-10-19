#r "packages/FAKE/tools/FakeLib.dll"
#r "packages/Fantomas/lib/FantomasLib.dll"

open Fake
open Fantomas.FakeHelpers
open Fantomas.FormatConfig

// Properties
let buildDir = "./build/"
let fantomasConfig =
    { FormatConfig.Default with
            PageWidth = 120
            ReorderOpenDeclaration = true }

Target "CheckCodeFormat" (fun _ ->
    !! "src/**/*.fs"
      |> checkCode fantomasConfig
)

Target "FormatCode" (fun _ ->
    !! "src/**/*.fs"
      |> formatCode fantomasConfig
      |> Log "Formatted files: "
)

RunTargetOrDefault "CheckCodeFormat"
