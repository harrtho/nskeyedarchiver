import AppKit
func createArchiveFiles(objects: NSArray, filename: String) {
    do {
        let archiver = NSKeyedArchiver(requiringSecureCoding: false)
        archiver.outputFormat = .xml
        for object in objects {
            archiver.encode(object)
        }
        archiver.finishEncoding()
        //let archivedXML = String(bytes: archiver.encodedData, encoding: .ascii)!

        let bin_archiver = NSKeyedArchiver(requiringSecureCoding: false)
        for object in objects {
            bin_archiver.encode(object)
        }
        bin_archiver.finishEncoding()

        let fullPath = URL(fileURLWithPath: filename + ".bin")
        try bin_archiver.encodedData.write(to: fullPath)
        let fullPath1 = URL(fileURLWithPath: filename + ".xml")
        try archiver.encodedData.write(to: fullPath1)
    } catch {
        print("Couldn't write file")
    }
}

createArchiveFiles(objects: [NSNumber(booleanLiteral: true)], filename: "../archiver/fixtures/boolean")
createArchiveFiles(objects: [NSNumber(booleanLiteral: true),2,3, "test", "test"], filename: "../archiver/fixtures/test")
